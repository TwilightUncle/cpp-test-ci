package buildenv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"dagger.io/dagger"
)

const (
	template_name     = "Dockerfile.template"
	template_path     = "./buildenv/templates/" + template_name
	output_root       = "./docker"
	output_filename   = "Dockerfile"
	guest_src_dirname = "/src"
	guest_vcpkg_dir   = "/opt/vcpkg"
)

func BuildDocker(ctx context.Context, client *dagger.Client, compiler string, version string, host_src_dirname string) (*dagger.Container, error) {
	if err := createDockerfile(compiler, version); err != nil {
		return nil, err
	}
	return client.Host().
		Directory("./"+getOutputDirPath(compiler, version)).
		DockerBuild().
		WithDirectory(guest_src_dirname, client.Host().Directory(host_src_dirname)).
		Sync(ctx)
}

func createDockerfile(compiler string, version string) error {
	// compilerに指定可能な値の制限
	if !slices.Contains([]string{"gcc", "clang"}, compiler) {
		return fmt.Errorf(
			"Invalid value \"%s\" specified for argument \"compiler\". Possible values for 'compiler' are \"gcc\" or \"clang\". ",
			compiler,
		)
	}

	// ディレクトリ作成(必要な場合のみ)
	if err := makeOutputDir(compiler, version); err != nil {
		return err
	}

	// ファイル生成(既存の場合上書き)
	fp, err := os.OpenFile(
		filepath.Join(getOutputDirPath(compiler, version), output_filename),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0777,
	)
	if err != nil {
		return err
	}
	defer fp.Close()

	// テンプレートの取得と結果ファイル生成
	tpl, err := template.New(template_name).ParseFiles(template_path)
	if err != nil {
		return err
	}
	return tpl.Execute(fp, GetCompilerInfo(compiler, version))
}

// 出力先ディレクトリが存在しない時のみ作成する
func makeOutputDir(compiler string, version string) error {
	out_dirpath := getOutputDirPath(compiler, version)
	f, err := os.Stat(out_dirpath)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(out_dirpath, 0777); err != nil {
			return err
		}
	} else if !f.IsDir() {
		return fmt.Errorf(
			"A file with the same name as the directory you are trying to create exists. filename: \"%s\"",
			out_dirpath,
		)
	}

	return nil
}

func GetCompilerInfo(compiler string, version string) map[string]string {
	c_compiler := ""
	cpp_compiler := ""
	install_pkgs := []string{}

	switch compiler {
	case "gcc":
		c_compiler = "gcc-" + version
		cpp_compiler = "g++-" + version
		install_pkgs = []string{c_compiler, cpp_compiler}
	case "clang":
		c_compiler = "clang-" + version
		cpp_compiler = "clang++-" + version
		install_pkgs = []string{c_compiler}
	// case "msvc":
	default:
		return nil
	}

	return map[string]string{
		"compiler_pkg_name": strings.Join(install_pkgs, " "),
		"c_compiler":        c_compiler,
		"cpp_compiler":      cpp_compiler,
		"src_dirname":       guest_src_dirname,
		"vcpkg_dir":         guest_vcpkg_dir,
	}
}

func getOutputDirPath(compiler string, version string) string {
	return filepath.Join(output_root, compiler, version)
}
