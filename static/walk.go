package static

import "path/filepath"

func Walk(path string, walkFunc filepath.WalkFunc) error {
	var explorer func(string) error
	explorer = func(path string) error {
		contents, err := AssetDir(path)
		if err != nil {
			panic(err)
		}
		for _, sub := range contents {
			subPath := path + "/" + sub
			fi, _ := AssetInfo(subPath)
			switch fi {
			case nil:
				if err := explorer(subPath); err != nil {
					return err
				}
			default:
				if err := walkFunc(subPath, fi, nil); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return explorer(path)
}