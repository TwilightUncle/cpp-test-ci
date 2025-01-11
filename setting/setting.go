package setting

import (
	"encoding/json"
	"os"
)

type Env struct {
	ProjectName     string
	TargetDirPath   string
	CompilerName    string
	CompilerVersion string
}

type projects struct {
	Projects []project `json:"projects"`
}

type project struct {
	Name          string     `json:"name"`
	TargetDirPath string     `json:"target_dir_path"`
	Compilers     []compiler `json:"compilers"`
}

type compiler struct {
	Name        string   `json:"name"`
	VersionList []string `json:"version_list"`
}

var Envs []Env

func Setup() error {
	envs, err := parseJson()
	Envs = envs
	return err
}

func parseJson() ([]Env, error) {
	raw_json, err := os.ReadFile("./setting.json")
	if err != nil {
		return []Env{}, err
	}

	var projects projects
	json.Unmarshal(raw_json, &projects)

	envs := []Env{}
	for _, project := range projects.Projects {
		for _, compiler := range project.Compilers {
			for _, v := range compiler.VersionList {
				envs = append(envs, Env{
					ProjectName:     project.Name,
					TargetDirPath:   project.TargetDirPath,
					CompilerName:    compiler.Name,
					CompilerVersion: v,
				})
			}
		}
	}

	return envs, nil
}
