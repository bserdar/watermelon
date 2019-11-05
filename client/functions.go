package client

var registeredFunctions = Functions{}

// Functions are used to keep track of the registered functions
type Functions map[string]func(*Session, []byte) ([]byte, error)

// Add adds a new function to the functions list
func (f Functions) Add(name string, function interface{}) {
	f[name] = Wrap(function)
}

// Export can be used to register a function as an anonymous variable
//
//   var _ = client.Export("myFunc",myFunc)
//
//   func myFunc() {
//   }
func Export(name string, function interface{}) int {
	registeredFunctions.Add(name, function)
	return 0
}
