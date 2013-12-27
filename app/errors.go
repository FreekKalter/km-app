package main

type Response struct {
	Body string
	Code int
}

var NotFound Response = Response{"404 page not found\n", 404}
var Ok Response = Response{"ok\n", 200}
var UnknownField Response = Response{"invalid fieldname\n", 400}
var NotParsable Response = Response{"could not parse request\n", 400}
var InvalidId Response = Response{"invalid id\n", 400}
var InvalidUrl Response = Response{"invalid url", 400}

var DbError Response = Response{"database eror", 500}
