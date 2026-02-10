package icons

import (
	"fmt"
	"strings"
)

// IconPack represents an icon pack with a name and loader URL.
type IconPack struct {
	Name string
	URL  string
}

// ParseIconPacks parses --iconPacks flags into IconPack structs.
// Format: @iconify-json/logos -> name=logos, url=https://unpkg.com/@iconify-json/logos/icons.json
func ParseIconPacks(packs []string) []IconPack {
	result := make([]IconPack, 0, len(packs))
	for _, pack := range packs {
		parts := strings.Split(pack, "/")
		name := parts[len(parts)-1]
		url := fmt.Sprintf("https://unpkg.com/%s/icons.json", pack)
		result = append(result, IconPack{Name: name, URL: url})
	}
	return result
}

// ParseIconPacksNamesAndUrls parses --iconPacksNamesAndUrls flags.
// Format: name#url
func ParseIconPacksNamesAndUrls(packs []string) []IconPack {
	result := make([]IconPack, 0, len(packs))
	for _, pack := range packs {
		idx := strings.Index(pack, "#")
		if idx < 0 {
			continue
		}
		name := pack[:idx]
		url := pack[idx+1:]
		result = append(result, IconPack{Name: name, URL: url})
	}
	return result
}

// GenerateIconPackJS generates JavaScript code to register icon packs with mermaid.
func GenerateIconPackJS(packs []IconPack) string {
	if len(packs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("mermaid.registerIconPacks([\n")
	for _, pack := range packs {
		sb.WriteString(fmt.Sprintf(`  {
    name: %q,
    loader: () => fetch(%q).then((res) => res.json()).catch(() => console.error("Failed to fetch icon: %s"))
  },
`, pack.Name, pack.URL, pack.Name))
	}
	sb.WriteString("]);\n")
	return sb.String()
}
