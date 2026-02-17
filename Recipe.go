//go:build ignore
package bake_recipe

import (
	"github.com/fezcode/gobake"
)

func Run(bake *gobake.Engine) error {
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		return err
	}
	// Ensure binary name matches revenge
	bake.Info.Name = "wilson-revenge"

	bake.Task("build", "Builds the binary for multiple platforms", func(ctx *gobake.Context) error {
		ctx.Log("Building %s v%s...", bake.Info.Name, bake.Info.Version)

		targets := []struct {
			os   string
			arch string
		}{
			{"linux", "amd64"},
			{"linux", "arm64"},
			{"windows", "amd64"},
			{"windows", "arm64"},
			{"darwin", "amd64"},
			{"darwin", "arm64"},
		}

		err := ctx.Mkdir("build")
		if err != nil {
			return err
		}

		for _, t := range targets {
			output := "build/" + bake.Info.Name + "-" + t.os + "-" + t.arch
			if t.os == "windows" {
				output += ".exe"
			}

			ctx.Env = []string{"CGO_ENABLED=0"}
			err := ctx.BakeBinary(t.os, t.arch, output)
			if err != nil {
				return err
			}
		}
		return nil
	})

	bake.Task("clean", "Removes build artifacts", func(ctx *gobake.Context) error {
		return ctx.Remove("build")
	})

	bake.Task("run-win", "Builds and runs the Windows version", func(ctx *gobake.Context) error {
		output := "build/" + bake.Info.Name + "-windows-amd64.exe"
		ctx.Log("Building for Windows...")
		ctx.Env = []string{"CGO_ENABLED=0"}
		if err := ctx.BakeBinary("windows", "amd64", output); err != nil {
			return err
		}
		ctx.Log("Running %s...", output)
		return ctx.Run(output)
	})

	return nil
}
