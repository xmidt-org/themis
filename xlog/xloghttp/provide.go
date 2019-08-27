package xloghttp

// ProvideStandardBuilders provides a standard set of logging fields for contextual handler logging.
// This function supplies the requestMethod, requestURI, and remoteAddr logging parameters.
func ProvideStandardBuilders() ParameterBuilders {
	return ParameterBuilders{
		Method("requestMethod"),
		URI("requestURI"),
		RemoteAddress("remoteAddr"),
	}
}
