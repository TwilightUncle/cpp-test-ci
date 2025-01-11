package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"dagger.io/dagger"

	"local-test-for-cpp/buildenv"
	"local-test-for-cpp/logging"
	"local-test-for-cpp/setting"
)

func main() {
	// read setting
	if err := setting.Setup(); err != nil {
		fmt.Println("Failed to read 'setting.json'")
		fmt.Println(err)
	}
	if err := build(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func build(ctx context.Context) error {
	fmt.Println("run test.")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	for _, env := range setting.Envs {
		compiler := env.CompilerName
		version := env.CompilerVersion

		err := logging.WithOutputLog(
			env.ProjectName+"-"+compiler+"-"+version+".log",
			func(logger *slog.Logger) error {
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
			},
		)

		if err != nil {
			return err
		}
	}

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
