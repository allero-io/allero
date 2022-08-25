package errorsHandler

type ErrorsHandler struct {
	errors []error
}

func New() *ErrorsHandler {
	return &ErrorsHandler{errors: []error{}}
}

func (eh *ErrorsHandler) Aggregate(errorsChan <-chan error) {
	errors := []error{}
	for e := range errorsChan {
		errors = append(errors, e)
	}

	eh.errors = errors
}

func (eh *ErrorsHandler) GetErrors() []error {
	return eh.errors
}
