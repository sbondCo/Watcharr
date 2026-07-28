<script lang="ts">
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import Icon from "@/lib/Icon.svelte";
	import { type AvailableAuthProviders } from "@/types";
	import { noAuthReq } from "@/lib/util/api";
	import { onMount } from "svelte";
	import { notify, unNotify } from "@/lib/util/notify";
	import { ReqerError } from "@/lib/util/fetch";

	let error: string | undefined = $state();
	let login = $state(true);
	let availableProviders: string[] = $state([]);
	let apHeader = $state(false);
	let apPlex = $state(false);
	let signupEnabled = $state(true);
	let useEmby = $state(false);
	let noAuto = $state(false);

	onMount(() => {
		if (localStorage.getItem("token")) {
			goto("/");
		}

		if (!error && page.url.searchParams.get("again")) {
			error = "Please Login Again";
		}
		if (page.url.searchParams.get("noAuto") == "1") {
			console.info(
				"login: Found noAuto param.. auto logins should be disabled now.",
			);
			noAuto = true;
		}

		processAvailableLoginMethods();
	});

	/**
	 * Get and process available login methods.
	 */
	async function processAvailableLoginMethods() {
		let r: AvailableAuthProviders;
		try {
			r = await noAuthReq.get<AvailableAuthProviders>("/auth/available");
		} catch {
			notify({
				type: "error",
				time: 10000,
				text: "Failed to get available login methods",
			});
			return;
		}
		if (!r) {
			console.log("AvailableAuth: no response data");
		}
		if (r.isInSetup) {
			console.log(
				"AvailableAuth: Server is in setup.. navigating to web setup page.",
			);
			goto("/setup");
		}
		availableProviders = r.available;
		apHeader = availableProviders?.includes("header");
		apPlex = availableProviders?.includes("plex");
		signupEnabled = r.signupEnabled;
		useEmby = r.useEmby;
		if (r.headerAuthAutoLogin && !noAuto) {
			console.log(
				"AvailableAuth: handling headerAuthAutoLogin.. calling proxyLogin automatically now.",
			);
			proxyLogin(true);
		}
	}

	function handleLogin(ev: SubmitEvent) {
		ev.preventDefault();
		const fd = new FormData(ev.target! as HTMLFormElement);
		const user = fd.get("username");
		const pass = fd.get("password");

		if (!user || !pass) {
			error = "Username and Password fields are required";
			return;
		}

		let customAuthEP = "";
		if ((ev.submitter as HTMLButtonElement)?.name === "jellyfin") {
			customAuthEP = "jellyfin";
		}

		const nid = notify({ text: "Logging in", type: "loading" });
		noAuthReq
			.post(`/auth${login ? `/${customAuthEP}` : "/register"}`, {
				username: user,
				password: pass,
			})
			.then((resp: any) => {
				if (resp?.token) {
					console.log("Received token... logging in.");
					localStorage.setItem("token", resp.token);
					if (useEmby) {
						localStorage.setItem("useEmby", "1");
					} else {
						localStorage.removeItem("useEmby");
					}
					goto("/");
					notify({ id: nid, text: `Welcome ${user}!`, type: "success" });
				}
			})
			.catch((err) => {
				error = ReqerError.getMsg(err, "Login failed");
				unNotify(nid);
			});
	}

	async function plexLogin() {
		try {
			const { preparePlexAuth, doPlexLogin, plexPinPoll } =
				await import("@/lib/util/plex");
			const p = preparePlexAuth();
			const pin = await doPlexLogin(p);
			plexPinPoll(pin, p, (err, token) => {
				if (err) {
					error = "Plex Auth Failed";
					console.error("Plex auth failed!", err);
					return;
				}
				const nid = notify({ text: "Logging in", type: "loading" });
				noAuthReq
					.post("/auth/plex", {
						token,
						clientIdentifier: p.clientId,
					})
					.then((resp: any) => {
						if (resp?.token) {
							console.log("Received token... logging in.");
							localStorage.setItem("token", resp.token);
							goto("/");
							notify({ id: nid, text: `Welcome!`, type: "success" });
						}
					})
					.catch((err) => {
						console.error("plexLogin: Fail", err);
						error = ReqerError.getMsg(err, "Login failed");
						notify({ id: nid, text: `Failed!`, type: "error" });
					});
			});
		} catch (err) {
			console.error("plexLogin: failed!", err);
			error = "Plex login failed";
		}
	}

	function proxyLogin(auto = false) {
		const nid = notify({ text: "Logging in", type: "loading" });
		noAuthReq
			.post(`/auth/proxy`)
			.then((resp: any) => {
				if (resp?.token) {
					console.log("Received token... logging in.");
					localStorage.setItem("token", resp.token);
					goto("/");
					notify({ id: nid, text: `Welcome!`, type: "success" });
				}
			})
			.catch((err) => {
				error = ReqerError.getMsg(err, "Login failed");
				if (auto) {
					notify({
						id: nid,
						text: `Automatic SSO Login Failed!`,
						type: "error",
					});
				} else {
					unNotify(nid);
				}
			});
	}
</script>

<div>
	<div class="inner">
		<h2>
			{#if login}
				Get Back In!
			{:else}
				Lucky You Found Us!
			{/if}
		</h2>

		{#if error}
			<span class="error">{error}!</span>
		{/if}

		<form onsubmit={handleLogin}>
			<label for="username">Username</label>
			<input type="text" name="username" placeholder="Username" />

			<label for="password">Password</label>
			<input type="password" name="password" placeholder="Password" />

			{#if login}
				<span class="login-with" style="font-weight: bold">Login With</span>
				<div class="login-btns">
					<button type="submit"><span class="watcharr">W</span>Watcharr</button>
					{#if availableProviders?.length > 0}
						{#if availableProviders.find((ap) => ap === "jellyfin")}
							{#if useEmby}
								<button type="submit" name="jellyfin" class="other">
									<Icon i="emby" wh={18} />
									emby
								</button>
							{:else}
								<button type="submit" name="jellyfin" class="other">
									<Icon i="jellyfin" wh={18} />
									jellyfin
								</button>
							{/if}
						{/if}
					{/if}
				</div>
				{#if apHeader || apPlex}
					<p style="font-weight: bold; font-size: 14px;">or</p>
					{#if apHeader}
						<div class="login-btns">
							<button
								type="button"
								name="proxy"
								class="proxy other"
								onclick={() => {
									proxyLogin();
								}}
							>
								<Icon i="lock-closed" wh={18} />Continue with Single Sign-On
							</button>
						</div>
					{/if}
					{#if apPlex}
						<div class="login-btns">
							<button
								type="button"
								onclick={() => {
									plexLogin();
								}}
								name="plex"
								class="plex other"
							>
								<Icon i="plex" wh={18} />Continue with Plex
							</button>
						</div>
					{/if}
				{/if}
			{:else}
				<div class="login-btns">
					<button type="submit">Sign Up</button>
				</div>
			{/if}
		</form>

		{#if signupEnabled}
			<button
				class="plain"
				onclick={() => {
					login = !login;
				}}
			>
				{#if login}
					Not a user?
				{:else}
					Already a user?
				{/if}
			</button>
		{/if}
	</div>
</div>

<style lang="scss">
	div,
	form {
		display: flex;
		flex-flow: column;
		align-items: center;
		gap: 10px;
		margin: 0 35px;
	}

	.inner,
	form {
		width: 100%;
		max-width: 400px;
	}

	.inner h2 {
		font-weight: normal;
	}

	label {
		align-self: flex-start;
		font-weight: bold;
	}

	span.login-with {
		font-size: 14px;
	}

	.login-btns {
		display: flex;
		flex-flow: row;
		gap: 10px;
		width: 100%;

		/* Hardcoded point for when main watcharr/jellyfin btns break. */
		@media screen and (max-width: 320px) {
			flex-wrap: wrap;
		}

		button {
			display: flex;
			flex-flow: row;
			gap: 10px;
			text-transform: capitalize;

			.watcharr {
				font-family: "Rampart One";
				font-size: 19px;
				line-height: 19px;
			}

			&.other {
				overflow: hidden;
				animation: 250ms ease otherbtn;

				@keyframes otherbtn {
					from {
						width: 0px;
					}
					to {
						width: 100%;
					}
				}
			}
		}
	}

	.error {
		display: flex;
		justify-content: center;
		width: 100%;
		padding: 10px;
		background-color: rgb(221, 48, 48);
		text-transform: capitalize;
		color: white;
	}
</style>
