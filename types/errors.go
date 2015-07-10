package types
import "fmt"


type ProjectNameError string

func (e ProjectNameError) Error() string {
	return string(e)
}

func InvalidProjectNameError(projectName ProjectName) ProjectNameError {
	return ProjectNameError(fmt.Sprintf("missing or bad ACKind (must be %#v)", projectName))
}