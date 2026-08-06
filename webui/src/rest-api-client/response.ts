export class ResponseError extends Error {
	// The HTTP status code of response
	httpStatusCode?: number;

	// The error code in the response
	errorCode?: number;

	// The stack trace of the server
	serverStackTrace?: string;

    name: string;

	/**
	 * Creates an instance of ResponseError.
	 * @param {string} message The error message
	 * @param {number} [errorCode] The error code which can be used by frontend to identify the actual error
	 * @param {number} [httpStatusCode] The HTTP status code the response should have
	 * @param {string} [stack] The stack trace
	 */
	constructor(
		message: string,
		options: { errorCode?: number; httpStatusCode?: number; stack?: string } = {},
        classicErrorOptions?: ErrorOptions
	) {
		super(message, classicErrorOptions);
		this.name = 'ResponseError';

		const { errorCode, httpStatusCode, stack, } = options;
		if (errorCode) {
			this.errorCode = errorCode;
		}
		if (httpStatusCode) {
			this.httpStatusCode = httpStatusCode;
		}
		if (stack) {
			this.serverStackTrace = stack;
		}
	}
}
