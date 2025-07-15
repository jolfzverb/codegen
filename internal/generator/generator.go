package generator

import (
	"strings"
)

func refIsExternal(ref string) bool {
	return !strings.HasPrefix(ref, "#")
}

func parseFilenameFromRef(ref string) string {
	parts := strings.SplitN(ref, "#", 2)
	if len(parts) != 2 {
		return ""
	}
	filename := parts[0]
	return filename
}

func (g *Generator) GetModelsImportForFile(filename string) string {
	return g.Opts.PackagePrefix + "/generated/" + g.GetModelName(filename) + "/" + g.GetModelName(filename) + "models"
}

func (g *Generator) GetHandlersImportForFile(filename string) string {
	return g.Opts.PackagePrefix + "/generated/" + g.GetModelName(filename)
}

func (g *Generator) ParseRefTypeName(ref string) (string, string) {
	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return "", ""
	}

	baseName := parts[len(parts)-1]

	if ref != "" && refIsExternal(ref) {
		filename := parseFilenameFromRef(ref)
		if filename == "" {
			return baseName, ""
		}
		g.YAMLFilesToProcess = append(g.YAMLFilesToProcess, filename)

		modelsImport := g.GetModelsImportForFile(filename)
		modelName := g.GetModelName(filename) + "models"

		return modelName + "." + baseName, modelsImport
	}

	return baseName, ""
}

func (g *Generator) GetCurrentModelsPackage() string {
	return g.PackageName + "models"
}
