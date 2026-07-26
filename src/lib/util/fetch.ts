import { goto } from "$app/navigation";
import { clearWatcharrData } from "../logout";
import { notify } from "./notify";

type ReqerParams = any;

/**
 * Request config, extending upon base RequestInit.
 */
export interface ReqerConfig extends RequestInit {
	/**
	 * URL Parameters.
	 *
	 * Example: `{ page: 1 }` will become &page=1 on the requests final url.
	 */
	params?: ReqerParams;
}

type ReqerConfigWithoutMethod = Omit<ReqerConfig, "method">;

export class ReqerError extends Error {
	constructor(
		public message: string,
		public body?: any,
		public response?: Response,
	) {
		super(message);
		this.name = "ReqerError";
	}

	/**
	 * If err is a ReqerError.
	 */
	static isReqerError(err: any) {
		return err instanceof ReqerError;
	}

	/**
	 * If err is a ReqerError and has a `body`.
	 */
	static withBody(
		err: any,
	): err is ReqerError & Required<Pick<ReqerError, "body">> {
		return this.isReqerError(err) && err.body;
	}

	/**
	 * Get error message from an error.
	 *
	 * If `err` is a ReqerError, use message from that instance, which can be
	 * the error returned from response body. Otherwise, standard error message
	 * is returned if one is availble.
	 *
	 * `act` should be the action that failed. It will be combined with the
	 * error message to create one readable string of `action: err`.
	 */
	static getMsg(err: any, act: string): string {
		// If response body has `error` property, is is a standard error
		// object that the Watcharr server returns. We don't want to return
		// a raw `body` if there is no expected `error` property incase it is
		// a random html error page or whatever (its better to show a pretty
		// error whenever possible, debugging errors are always in console),
		// if there is no `error`, we try the `err.message`.
		let msg = "You've encountered an extremely unguarded error!";
		if (ReqerError.withBody(err) && err.body.error) {
			msg = err.body.error;
		} else if (err.message) {
			msg = err.message;
		}
		return `${act}: ${msg}`;
	}
}

/**
 * Reqer request response object.
 * The *Whole functions return this whole object, which includes
 * request details (like status), along with the response body.
 */
export interface ReqerResponse<T> {
	/**
	 * Parsed body.
	 */
	body: T;
	/**
	 * Response status code.
	 */
	status: number;
	/**
	 * Response headers.
	 */
	headers: Headers;
}

/**
 * A simple Fetch API wrapper.
 *
 * Throws when response.status is not 2xx.
 *
 * Has nice `get`, `post`, etc methods that simply return response body. Along
 * with their `*Whole` counterpart methods that return a `ReqerResponse` object
 * for more involved logic that needs to look at status code or headers.
 */
export class Reqer {
	constructor(
		/**
		 * BaseURL for all requests.
		 */
		private baseUrl: string,
		/**
		 * If the requests should add an Authorization header
		 * with users Watcharr token.
		 */
		private watcharrAuthed: boolean,
	) {}

	private buildBaseUrl(b: string): string {
		if (!b.endsWith("/")) {
			b += "/";
		}
		return b;
	}

	private buildUrlPath(p: string): string {
		// Removes any slashes or "." from start of path.
		return p.replace(/^[.\/\\]+/, "");
	}

	private buildUrl(p: string, params?: ReqerParams): URL {
		try {
			// URL builds the url string "properly", but I don't want that for
			// this use case, so we are pre-processing the path and baseUrl to
			// skirt the relative resolving of the paths.
			// See: https://developer.mozilla.org/en-US/docs/Web/API/URL_API/Resolving_relative_references
			const url = new URL(
				this.buildUrlPath(p),
				this.buildBaseUrl(this.baseUrl),
			);
			if (params && typeof params === "object") {
				for (const k in params) {
					if (!Object.hasOwn(params, k)) continue;
					const el = params[k];
					if (el) {
						// We use append on url.searchParams instead of
						// overwriting it with a new object incase it has any
						// existing params parsed from the `p`ath.
						url.searchParams.append(k, String(el));
					}
				}
			}
			return url;
		} catch (err) {
			console.error("reqer->buildUrl: Failed!", err);
			throw new ReqerError("building request url failed");
		}
	}

	private prepareRequestBody(
		data: any,
		headers: Headers,
	): BodyInit | undefined {
		if (!data) {
			return;
		}

		// The browser can handle these.
		if (
			data instanceof FormData ||
			data instanceof Blob ||
			data instanceof URLSearchParams ||
			data instanceof ArrayBuffer ||
			ArrayBuffer.isView(data)
		) {
			return data as BodyInit;
		}

		// Already encoded, so return it as is.
		if (typeof data === "string") {
			return data;
		}

		// Stringify json.
		headers.append("Content-Type", "application/json");
		return JSON.stringify(data);
	}

	private async parseResponseBody(res: Response) {
		const contentType = res.headers.get("Content-Type");
		if (!contentType) {
			throw new ReqerError("response has no content-type", res);
		}
		if (contentType.includes("application/json")) {
			return await res.json();
		} else if (contentType.includes("text/plain")) {
			return await res.text();
		}
		throw new ReqerError("response includes an unsupported content-type", res);
	}

	async do<T>(p: string, cfg?: ReqerConfig): Promise<ReqerResponse<T>> {
		try {
			const headers = new Headers(cfg?.headers);

			if (this.watcharrAuthed) {
				const token = localStorage.getItem("token");
				if (!token) {
					console.error("No token, going to login.");
					goto("/login?again=1");
					throw new ReqerError("No auth token found");
				}
				headers.append("Authorization", token);
			}

			const reqBody = this.prepareRequestBody(cfg?.body, headers);

			const res = await fetch(this.buildUrl(p, cfg?.params), {
				...cfg,
				body: reqBody,
				headers,
			});

			let resBody = await this.parseResponseBody(res);

			if (!res.ok) {
				throw new ReqerError(
					`request failed with ${res.status} ${res.statusText}`,
					resBody,
					res,
				);
			}

			return {
				body: resBody,
				status: res.status,
				headers: res.headers,
			};
		} catch (err) {
			console.error("Reqer->do: Errored!", err);
			if (err instanceof ReqerError) {
				if (this.watcharrAuthed && err.response?.status === 401) {
					console.error("Recieved 401 response, going to login.");
					notify({ text: "Request Authorization Failed!", type: "error" });
					clearWatcharrData();
					goto("/login?again=1");
				}
				throw err;
			} else if (err instanceof Error) {
				throw new ReqerError(err.message);
			}
			throw new ReqerError("request failed");
		}
	}

	/**
	 * GET request, returning parsed response body.
	 */
	async get<T>(p: string, cfg?: ReqerConfigWithoutMethod): Promise<T> {
		return (await this.do<T>(p, { ...cfg, method: "GET" })).body;
	}

	/**
	 * Same as `get()`, except the whole ReqerResponse object is returned.
	 */
	async getWhole<T>(
		p: string,
		cfg?: ReqerConfigWithoutMethod,
	): Promise<ReqerResponse<T>> {
		return await this.do<T>(p, { ...cfg, method: "GET" });
	}

	/**
	 * POST request, returning parsed response body.
	 */
	async post<T>(
		p: string,
		data?: any,
		cfg?: Omit<ReqerConfigWithoutMethod, "body">,
	): Promise<T> {
		return (await this.do<T>(p, { ...cfg, method: "POST", body: data })).body;
	}

	/**
	 * Same as `post()`, except the whole ReqerResponse object is returned.
	 */
	async postWhole<T>(
		p: string,
		data?: any,
		cfg?: Omit<ReqerConfigWithoutMethod, "body">,
	): Promise<ReqerResponse<T>> {
		return await this.do<T>(p, { ...cfg, method: "POST", body: data });
	}

	/**
	 * PUT request, returning parsed response body.
	 */
	async put<T>(
		p: string,
		data?: any,
		cfg?: ReqerConfigWithoutMethod,
	): Promise<T> {
		return (await this.do<T>(p, { ...cfg, method: "PUT", body: data })).body;
	}

	/**
	 * Same as `put()`, except the whole ReqerResponse object is returned.
	 */
	async putWhole<T>(
		p: string,
		data?: any,
		cfg?: ReqerConfigWithoutMethod,
	): Promise<ReqerResponse<T>> {
		return await this.do<T>(p, { ...cfg, method: "PUT", body: data });
	}

	/**
	 * DELETE request, returning parsed response body.
	 */
	async delete<T>(p: string, cfg?: ReqerConfigWithoutMethod): Promise<T> {
		return (await this.do<T>(p, { ...cfg, method: "DELETE" })).body;
	}

	/**
	 * Same as `delete()`, except the whole ReqerResponse object is returned.
	 */
	async deleteWhole<T>(
		p: string,
		cfg?: ReqerConfigWithoutMethod,
	): Promise<ReqerResponse<T>> {
		return await this.do<T>(p, { ...cfg, method: "DELETE" });
	}
}
