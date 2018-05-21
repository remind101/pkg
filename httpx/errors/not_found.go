package errors

func NotFoundError() *notFound {
	return &notFound{}
}

type notFound struct{}

func (n *notFound) Error() string {
	return "Not found"
}

func (n *notFound) StatusCode() int {
	return 404
}
