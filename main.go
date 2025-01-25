package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"dagger.io/dagger"

	"cpp-test-ci/buildenv"
	"cpp-test-ci/logging"
	"cpp-test-ci/setting"
)

func main() {
	// read setting
	if err := setting.Setup(); err != nil {
		fmt.Println("Failed to read 'setting.json'")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// run workflow
	errors := buildAll(context.Background())
	for _, err := range errors {
		fmt.Fprintln(os.Stderr, err)
	}
	if len(errors) > 0 {
		os.Exit(1)
	}
}

func buildAll(ctx context.Context) []error {
	fmt.Println("run test.")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return []error{err}
	}
	defer client.Close()

	// 各環境毎のテスト実行までは平行
	result := make(chan error)
	for _, env := range setting.Envs {
		go func(e setting.Env) {
			compiler := e.CompilerName
			version := e.CompilerVersion

			// 処理系毎に出力するログも分け、結果はチャネルに格納
			result <- logging.WithOutputLog(
				e.ProjectName+"-"+compiler+"-"+version+".log",
				func(logger *slog.Logger) error {
					return build(e, ctx, client, logger)
				},
			)
		}(env)
	}

	// 各ルーチンの結果収集
	errors := []error{}
	for range setting.Envs {
		if err := <-result; err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// 各処理系毎のbuild
func build(env setting.Env, ctx context.Context, client *dagger.Client, logger *slog.Logger) error {
	compiler := env.CompilerName
	version := env.CompilerVersion

	logger.Info(fmt.Sprintf("## start. target_project: %s, compiler: %s, compiler_version: %s", env.ProjectName, compiler, version))
	logger.Info("== building dockerfile ================")
	source, err := buildenv.BuildDocker(ctx, client, compiler, version, env.TargetDirPath)
	if err != nil {
		logger.Error("Error occured. :" + err.Error())
		return err
	}

	env_info := buildenv.GetCompilerInfo(compiler, version)

	logger.Info("== configure project ================")
	configured, msg, err := configureCmake(ctx, source, env_info)
	if err != nil {
		logger.Error("Error occured. :" + err.Error())
		return err
	}
	logger.Info(msg)

	logger.Info("== build project ================")
	built, msg2, err := buildBySource(ctx, configured, env_info)
	if err != nil {
		logger.Error("Error occured. :" + err.Error())
		return err
	}
	logger.Info(msg2)

	// run test
	logger.Info("== run test =============")
	_, msg3, err := runTest(ctx, built, env_info)
	if err != nil {
		logger.Error("Error occured. :" + err.Error())
		return err
	}
	logger.Info(msg3)

	return nil
}

// CMakeLists.txtを参照の上、プロジェクトを構築する
func configureCmake(ctx context.Context, container *dagger.Container, env_info map[string]string) (*dagger.Container, string, error) {
	configured, err := container.WithExec([]string{
		"cmake",
		"-B", env_info["src_dirname"] + "/build",
		"-S", env_info["src_dirname"],
		"-DCMAKE_BUILD_TYPE=Release",
		fmt.Sprintf("-DCMAKE_TOOLCHAIN_FILE=%s/scripts/buildsystems/vcpkg.cmake", env_info["vcpkg_dir"]),
	}).Sync(ctx)

	if err != nil {
		return configured, "", err
	}

	out, err := configured.Stdout(ctx)
	return configured, out, err
}

// プロジェクトのソースをビルドし、バイナリ生成
func buildBySource(ctx context.Context, container *dagger.Container, env_info map[string]string) (*dagger.Container, string, error) {
	built, err := container.WithExec([]string{
		"cmake",
		"--build", env_info["src_dirname"] + "/build",
		"--config", "Release",
	}).Sync(ctx)

	if err != nil {
		return built, "", err
	}

	out, err := built.Stdout(ctx)
	return built, out, err
}

func runTest(ctx context.Context, container *dagger.Container, env_info map[string]string) (*dagger.Container, string, error) {
	executed, err := container.
		WithExec([]string{
			"ctest",
			"-C", "Release",
			"--test-dir", env_info["src_dirname"] + "/build",
			"--output-on-failure",
		}).Sync(ctx)

	if err != nil {
		return executed, "", err
	}

	out, err := executed.Stdout(ctx)
	return executed, out, err
}
