package zepository

import (
	"fmt"
	"github.com/asaskevich/govalidator"
)

var references []ReferenceInterface

func init() {
	govalidator.TagMap["manala_repository_description"] = govalidator.Validator(func(str string) bool {
		return govalidator.Matches(str, "^[a-z.-]{1,64}$")
	})
}

/*************/
/* Reference */
/*************/

type ReferenceInterface interface {
	GetRepositorySource() string
}

func AddReference(reference ReferenceInterface) {
	references = append(references, reference)
	fmt.Printf("%#v\n", references)
}
