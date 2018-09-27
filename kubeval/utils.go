package kubeval

// Based on https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang
// We unmarshal yaml into a value of type interface{},
// go through the result recursively, and convert each encountered
// map[interface{}]interface{} to a map[string]interface{} value
// required to marshall to JSON.
func convertToStringKeys(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertToStringKeys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertToStringKeys(v)
		}
	}
	return i
}

func unpackGzipFile(gzFilePath, dstFilePath string) (int64, error) {
    gzFile, err := os.Open(gzFilePath)
    if err != nil {
        return 0, fmt.Errorf("Failed to open file %s for unpack: %s", gzFilePath, err)
    }
    dstFile, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
    if err != nil {
        return 0, fmt.Errorf("Failed to create destination file %s for unpack: %s", dstFilePath, err)
    }

    ioReader, ioWriter := io.Pipe()

    go func() { // goroutine leak is possible here
        gzReader, _ := gzip.NewReader(gzFile)
        // it is important to close the writer or reading from the other end of the
        // pipe or io.copy() will never finish
        defer func(){
            gzFile.Close()
            gzReader.Close()
            ioWriter.Close()
        }()

        io.Copy(ioWriter, gzReader)
    }()

    written, err := io.Copy(dstFile, ioReader)
    if err != nil {
        return 0, err // goroutine leak is possible here
    }
    ioReader.Close()
    dstFile.Close()

    return written, nil
}
