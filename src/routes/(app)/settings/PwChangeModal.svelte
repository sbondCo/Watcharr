<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Modal from "@/lib/Modal.svelte";
	import type { ChangePasswordForm } from "@/types";
	import { changeUserPassword } from "$lib/util/api";

	interface Props {
		userName: string | undefined;
		onClose: () => void;
		changepswd?: ChangePasswordForm;
	}

	let {
		userName,
		onClose,
		changepswd = $bindable({
			currentPassword: "",
			newPassword: "",
			reEnteredNewPassword: "",
		}),
	}: Props = $props();

	let error: string | undefined = $state();
	let errs: string[] = [];

	function checkForm() {
		errs = [];
		console.log(
			"Change Password Form - Status: Checking if any form inputs are missing",
		);
		if (!changepswd.currentPassword) {
			errs.push(t("password.current"));
		}
		if (!changepswd.newPassword) {
			errs.push(t("password.new"));
		}
		if (!changepswd.reEnteredNewPassword) {
			errs.push(t("password.reEnteredNew"));
		}
		if (errs.length > 0) {
			error = t("password.missingParams", { params: errs.join(", ") });
			console.log(
				`Change Password Form - Error: following form inputs are missing:\n${errs.join("\n")}`,
			);
		} else {
			console.log("Change Password Form - Status: All form inputs are present");
			checkFormPasswordsMatch();
		}
	}

	function checkFormPasswordsMatch() {
		console.log(
			"Change Password Form - Status: Checking if new password and re-entered new password match",
		);
		if (changepswd.newPassword !== changepswd.reEnteredNewPassword) {
			error = t("password.doNotMatch");
			console.log(
				"Change Password Form - Error: New password and re-entered new password do not match",
			);
		} else {
			console.log(
				"Change Password Form - Status: New password and re-entered new password match",
			);
			error = "";
		}
	}

	function handleSubmit(ev: SubmitEvent) {
		ev.preventDefault();
		checkForm();
		if (!error) {
			console.log(
				"Change Password Form - Status: Form inputs on the frontend are valid - sending request to change password to the server",
			);
			const fd = new FormData(ev.target! as HTMLFormElement);
			changepswd.currentPassword = fd.get("current-password") as string;
			changepswd.newPassword = fd.get("new-password") as string;
			changeUserPassword(
				changepswd.currentPassword,
				changepswd.newPassword,
				(errMsg) => {
					if (errMsg) {
						error = errMsg;
						return;
					}
					console.log(
						"Change Password Form - Status: Password changed successfully",
					);
					onClose();
				},
			);
		}
	}
</script>

<Modal
	title={t("settings.changePassword")}
	desc={t("password.desc", { user: userName })}
	{onClose}
>
	{#if error}
		<span class="error">{error}!</span>
	{/if}
	<form onsubmit={handleSubmit}>
		<div class="form-input-container">
			<div class="form-input">
				<!--Hiding username info as it is still useful to password managers-->
				<!--https://www.chromium.org/developers/design-documents/create-amazing-password-forms/#use-hidden-fields-for-implicit-information-->
				<label for="username" id="username-label">{t("login.username")}</label>
				<input
					type="text"
					name="username"
					autocomplete="username"
					id="username-input"
					value={userName}
				/>
			</div>
			<div class="form-input">
				<label for="current-password">{t("password.current")}</label>
				<input
					type="password"
					name="current-password"
					placeholder={t("password.currentPlaceholder")}
					autocomplete="current-password"
					bind:value={changepswd.currentPassword}
				/>
			</div>
			<div class="form-input">
				<label for="new-password">{t("password.new")}</label>
				<input
					type="password"
					name="new-password"
					placeholder={t("password.newPlaceholder")}
					autocomplete="new-password"
					bind:value={changepswd.newPassword}
				/>
			</div>
			<div class="form-input">
				<label for="re-entered-new-password">{t("password.reEnterNew")}</label>
				<input
					type="password"
					name="re-entered-new-password"
					placeholder={t("password.reEnterNewPlaceholder")}
					autocomplete="new-password"
					bind:value={changepswd.reEnteredNewPassword}
				/>
			</div>
			<button type="submit">{t("settings.changePassword")}</button>
		</div>
	</form>
</Modal>

<style lang="scss">
	#username-label,
	#username-input {
		display: none;
	}

	.error {
		position: sticky;
		top: 0;
		display: flex;
		justify-content: center;
		width: 100%;
		padding: 10px;
		background-color: rgb(221, 48, 48);
		text-transform: capitalize;
		color: white;
		margin-bottom: 15px;
	}

	.form-input-container {
		display: flex;
		flex-flow: column;
		gap: 20px;
		margin: 0 15px;
	}

	.form-input-container > button {
		margin-left: auto;
		width: max-content;
	}
</style>
