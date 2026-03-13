package commitlog

import "os"

func expectFiles(path string, expected []string) (bool, error) {
	filesSet := make(map[string]struct{}, len(expected))
	for _, file := range expected {
		filesSet[file] = struct{}{}
	}

	gotSet := make(map[string]struct{})
	dirs, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		name := dir.Name()
		if _, ok := filesSet[name]; !ok {
			return false, nil // unexpected file found
		}
		gotSet[name] = struct{}{}
	}

	return len(gotSet) == len(filesSet), nil
}
