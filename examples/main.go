package main

import (
	"github.com/bb-pencil/bbloger"
)

type E struct {
	str string
}

func (e E) Error() string {
	return e.str
}

func main() {
	bbloger.SetVerbosity(1)
	//bblog := bbloger.New(log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile))
	bblog := bbloger.New(nil)
	bblog = bblog.WithName("MyName").WithValues("user", "you")
	bblog.Info("hello", "val1", 1, "val2", map[string]int{"k": 1})
	bblog.V(1).Info("you should see this")
	bblog.V(1).V(1).Info("you should not see this")
	bblog.Error(nil, "oh oh", "trouble", true, "reason", []float64{0.1, 0.11, 3,14})
	bblog.Error(E{"an error occurred"}, "goodbye", "code", -1)
}
