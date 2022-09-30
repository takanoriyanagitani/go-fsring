package fsring

type Read func(filename string) (data []byte, e error)
