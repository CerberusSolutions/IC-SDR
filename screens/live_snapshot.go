package screens

import "os"

// replaceLiveSnapshot publishes a complete JSON snapshot on Windows too.
// os.Rename does not replace an existing destination there, so remove the old
// snapshot immediately before promoting the fully-written temporary file.
func replaceLiveSnapshot(path string, data []byte) error {
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0644); err != nil {
		return err
	}
	_ = os.Remove(path)
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}
