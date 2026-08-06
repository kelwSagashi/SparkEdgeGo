type GenericValue = string | number | boolean | object | null | undefined | GenericValue[];
import axios from 'axios';
import type { AxiosRequestConfig, Method, RawAxiosRequestHeaders } from 'axios';
import { ResponseError } from './response';

export const BROWSER_ID_STORAGE_KEY = 'spark-edge-browserId'

const getBrowserId = () => {
	let browserId = localStorage.getItem(BROWSER_ID_STORAGE_KEY);
	if (!browserId) {
		browserId = crypto.randomUUID();
		localStorage.setItem(BROWSER_ID_STORAGE_KEY, browserId);
	}
	return browserId;
};

const legacyParamSerializer = (params: Record<string, any>) =>
	Object.keys(params)
		.filter((key) => params[key] !== undefined)
		.map((key) => {
			if (Array.isArray(params[key])) {
				return params[key].map((v: string) => `${key}[]=${encodeURIComponent(v)}`).join('&');
			}
			if (typeof params[key] === 'object') {
				params[key] = JSON.stringify(params[key]);
			}
			return `${key}=${encodeURIComponent(params[key])}`;
		})
		.join('&');

export class MfaRequiredError extends ResponseError {
	constructor() {
		super('MFA is required to access this resource. Please set up MFA in your user settings.');
		this.name = 'MfaRequiredError';
	}
}

export async function request(config: {
	method: Method;
	baseURL: string;
	endpoint: string;
	headers?: RawAxiosRequestHeaders;
	data?: GenericValue | GenericValue[];
	withCredentials?: boolean;
}) {
	const { method, baseURL, endpoint, headers, data } = config;
	const options: AxiosRequestConfig = {
		method,
		url: endpoint,
		baseURL,
		headers: headers ?? {},
	};

    if (baseURL.startsWith('/')) {
		options.headers!['browser-id'] = getBrowserId();
	}
	
	if (
		import.meta.env.NODE_ENV !== 'production' &&
		!baseURL.includes('api.spark-edge.io') &&
		!baseURL.includes('spark-edge.cloud')
	) {
		options.withCredentials = options.withCredentials ?? true;
	}
	if (['POST', 'PATCH', 'PUT'].includes(method)) {
		options.data = data;
	} else if (data) {
		options.params = data;
		options.paramsSerializer = legacyParamSerializer;
	}

	try {
		const response = await axios.request(options);
		return response.data;
	} catch (error: any) {
		if (error.message === 'Network Error') {
			throw new ResponseError("Can't connect to spark-edge.", {
				errorCode: 999,
			});
		}

		const errorResponseData = error.response?.data;
		if (errorResponseData?.mfaRequired === true) {
			throw new MfaRequiredError();
		}
		if (errorResponseData?.message !== undefined) {
			if (errorResponseData.name === 'NodeApiError') {
				errorResponseData.httpStatusCode = error.response.status;
				throw errorResponseData;
			}
            
			throw new ResponseError(errorResponseData.message, {
				errorCode: errorResponseData.code,
				httpStatusCode: error.response.status,
				stack: errorResponseData.stack,
			});
		}

		throw error;
	}
}
