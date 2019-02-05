package controller

import (
	"github.com/elastic/k8s-operators/operators/pkg/controller/common/operator"
	"github.com/elastic/k8s-operators/operators/pkg/controller/stack"
)

func init() {
	Register(operator.ApplicationOperator, stack.Add)
}
