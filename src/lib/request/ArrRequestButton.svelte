<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import type {
		ArrDetailsResponse,
		ArrInfoResponse,
		ArrRequestResponse,
		ArrRequestStatus,
		ContentType,
	} from "@/types";
	import { onMount } from "svelte";
	import Icon from "../Icon.svelte";
	import tooltip from "../actions/tooltip";
	import { msToAmountsOfTime } from "../util/helpers";
	import { notify } from "../util/notify";
	import { req } from "../util/api";
	import { ReqerError } from "../util/fetch";

	interface Props {
		type: ContentType;
		tmdbId: number;
		openRequestModal: () => void;
	}

	let { type, tmdbId, openRequestModal }: Props = $props();

	let existingRequest: ArrRequestResponse | undefined;
	let info: ArrInfoResponse | undefined;
	// The status, extra added string types are set in this file.
	let status:
		| ArrDetailsResponse
		| "available"
		| "requested"
		| ArrRequestStatus
		| undefined = $state();
	let estimatedCompletionIn: string | undefined = $state();

	async function getInfo() {
		try {
			if (!existingRequest) {
				console.warn("getInfo called before existingRequest exists.");
				return;
			}
			if (
				existingRequest.status !== "APPROVED" &&
				existingRequest.status !== "AUTO_APPROVED" &&
				existingRequest.status !== "FOUND"
			) {
				console.debug(
					"getInfo: Request is not in an approved state.. not continuing.",
				);
				return;
			}
			const resp = await req.get<ArrInfoResponse>(
				`/arr/${type === "movie" ? "rad" : "son"}/info/${existingRequest.id}`,
			);
			if (resp) {
				info = resp;
				if (info.hasFile) {
					status = "available";
				} else {
					getStatus();
				}
			}
		} catch (err) {
			if (
				ReqerError.withBody(err) &&
				err.response?.status === 404 &&
				err.hasErrorInBody() &&
				err.body.error === "request deleted"
			) {
				return;
			}
			console.error("ArrRequestButton: getInfo failed!", err);
			notify({
				text: t("request.failedGettingInfo", {
					service: type === "movie" ? "Radarr" : "Sonarr",
				}),
				type: "error",
			});
		}
	}

	async function getStatus(ev: MouseEvent | undefined = undefined) {
		try {
			if (!existingRequest) {
				console.warn("getStatus called before existingRequest exists.");
				return;
			}
			const statusResp = await req.getWhole<ArrDetailsResponse>(
				`/arr/${type === "movie" ? "rad" : "son"}/status/${existingRequest.serverName}/${existingRequest.arrId}`,
			);
			if (statusResp.status === 204) {
				status = "requested";
			} else if (statusResp.body) {
				status = statusResp.body;
				const estMs =
					new Date(status.estimatedCompletionTime).getTime() -
					new Date(Date.now()).getTime();
				const est = msToAmountsOfTime(estMs);
				if (est.days > 0) {
					estimatedCompletionIn = t("request.days", { count: est.days });
				} else if (est.hours > 0) {
					estimatedCompletionIn = t("request.hours", { count: est.hours });
				} else if (est.minutes > 0) {
					estimatedCompletionIn = t("request.minutes", { count: est.minutes });
				} else if (est.seconds > 0) {
					estimatedCompletionIn = t("request.seconds", { count: est.seconds });
				} else {
					estimatedCompletionIn = undefined;
				}
			}
			if (ev?.target) {
				setTimeout(() => {
					(ev?.target as HTMLButtonElement)?.blur();
				}, 250);
			}
		} catch (err) {
			console.error("ArrRequestButton: getStatus failed!", err);
			notify({
				text: t("request.failedFetchingStatus", {
					service: type === "movie" ? "Radarr" : "Sonarr",
				}),
				type: "error",
			});
		}
	}

	async function lookForExisting() {
		try {
			const existingRequestResp = await req.get<ArrRequestResponse>(
				`/arr/${type === "movie" ? "rad" : "son"}/request/${tmdbId}`,
			);
			if (existingRequestResp && existingRequestResp.arrId) {
				existingRequest = existingRequestResp;
				getInfo();
			} else if (existingRequestResp) {
				// If no arrId, use status in request (pending, denied, etc)
				console.log(
					"No arrId in request resp.. using request status for btn status if set.",
				);
				if (existingRequestResp.status) {
					status = existingRequestResp.status;
				}
			}
		} catch (err) {
			console.error("ArrRequestButton: lookForExisting failed!", err);
			notify({
				text: t("request.failedLookingForExisting"),
				type: "error",
			});
		}
	}

	// Used by Request modals to set existing request data in this component from their request response.
	export function setExistingRequest(r: ArrRequestResponse) {
		if (existingRequest) {
			console.error(
				"ArrRequestButton: setExistingRequest: Existing request has already been defined! Not continuing.",
			);
			return;
		}
		if (!r) {
			console.error(
				"ArrRequestButton: setExistingRequest: No response passed in.",
				r,
			);
			return;
		}
		console.debug("ArrRequestButton: setExistingRequest: Running..", r);
		existingRequest = r;
		if (existingRequest?.status) {
			status = existingRequest?.status;
		}
		getInfo();
	}

	onMount(() => {
		lookForExisting();
	});
</script>

{#if typeof status === "object" || status === "requested"}
	<button
		onclick={getStatus}
		use:tooltip={{
			text:
				status === "requested"
					? t("request.waitingForApproval")
					: status.status === "downloading"
						? estimatedCompletionIn
							? t("request.doneIn", { time: estimatedCompletionIn })
							: t("request.estimationUnavailable")
						: status.status === "paused"
							? t("request.downloadPaused")
							: "",
			pos: "bot",
		}}
	>
		<div><Icon i="refresh" /></div>
		<span>
			{#if status === "requested"}
				{t("request.status.requested")}
			{:else}
				{status.status}
				{status.status === "downloading"
					? status?.progress
						? `(${status?.progress}%)`
						: ""
					: ""}
			{/if}
		</span>
	</button>
{:else if status === "available" || status === "FOUND"}
	<button disabled>{t("request.status.available")}</button>
{:else if status === "PENDING"}
	<button disabled>{t("request.status.pending")}</button>
{:else if status === "APPROVED"}
	<button disabled>{t("request.status.approved")}</button>
{:else if status === "AUTO_APPROVED"}
	<button disabled>{t("request.status.autoApproved")}</button>
{:else if status === "DENIED"}
	<button disabled>{t("request.status.denied")}</button>
{:else}
	<button onclick={openRequestModal}>{t("request.request")}</button>
{/if}

<style lang="scss">
	button {
		max-width: fit-content;
		text-transform: capitalize;
		position: relative;

		& > div {
			display: none;
			position: absolute;
			left: 50%;
			top: 50%;
			transform: translate(-50%, -50%);
		}

		&:hover:not(:focus) {
			& > div {
				display: block;
			}

			& > span {
				visibility: hidden;
			}
		}
	}
</style>
