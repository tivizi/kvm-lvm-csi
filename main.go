package main

func main() {
	driver, err := NewDriver()
	if err != nil {
		panic(err)
	}
	driver.Run()
}
