package apppackage

import (
	"fmt"
	"testing"
)

func TestJarEnc(t *testing.T) {

	encrypt, err := jarEncrypt("/home/slx/00-Applications/03-java/01java8/jdk1.8.0_181/bin/java",
		"/home/slx/work/10-可执行jar加密打包运行/common-utils/apppackage/jar-enc.jar",
		"/home/slx/work/10-可执行jar加密打包运行/common-utils/apppackage/demo.jar",
		"/home/slx/work/10-可执行jar加密打包运行/common-utils/apppackage/jre/target.jar", "sdgf1564345", nil, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Printf("%s\n", encrypt)
}
