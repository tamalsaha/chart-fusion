package subpkg_test

import (
	"fmt"
	"github.com/gobuffalo/flect"
	"strings"
	"testing"
)

func TestFlect(t *testing.T) {
	in := "kubedb.com"
	v := strings.Replace(in, ".", "_", -1)
	//s := flect.Camelize(v)
	//fmt.Println(s)

	fmt.Println(flect.Dasherize(v))
	// fmt.Println(flect.Capitalize(v))
	fmt.Println(flect.Pascalize(v))
	fmt.Println(flect.Camelize(v))
}
