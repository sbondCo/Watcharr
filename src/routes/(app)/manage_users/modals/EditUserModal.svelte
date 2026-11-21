<script lang="ts">
	import { t } from "@/lib/i18n/index.svelte";
	import Checkbox from "@/lib/Checkbox.svelte";
	import DropDown from "@/lib/DropDown.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { req } from "@/lib/util/api";
	import { ReqerError } from "@/lib/util/fetch";
	import { userHasPermission } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import { UserPermission, UserType, type ManagedUser } from "@/types";

	interface UpdateUserRequest {
		permissions?: number;
		type?: UserType;
	}

	interface Props {
		user: ManagedUser;
		onClose: () => void;
	}

	let { user = $bindable(), onClose }: Props = $props();

	let error: string | undefined = $state();

	// Things we have changed
	let changedPerms = false;
	let originalUser = structuredClone(user);

	async function save() {
		const changedType = user.type !== originalUser.type;
		// If nothing changed.. error
		if (!changedPerms && !changedType) {
			error = t("manageUsers.edit.nothingChanged");
			return;
		}
		if (!error) {
			try {
				const toUpdate: UpdateUserRequest = {};
				if (changedPerms) {
					toUpdate["permissions"] = user.permissions;
				}
				if (changedType) {
					toUpdate["type"] = user.type;
				}
				await req.post(`/server/users/${user.id}`, toUpdate);
				notify({
					type: "success",
					text: t("manageUsers.edit.changesSaved"),
				});
				onClose();
			} catch (err) {
				console.error("Failed to save user!", err);
				error = ReqerError.getMsg(err, t("manageUsers.edit.failedToSave"));
			}
		}
	}

	function userTogglePermission(perm: UserPermission) {
		user.permissions ^= perm;
		changedPerms = true;
	}
</script>

<Modal
	title={t("manageUsers.edit.title")}
	desc={t("manageUsers.edit.desc", { user: user.username })}
	maxWidth="500px"
	{onClose}
>
	{#if error}
		<span class="error">{error}!</span>
	{/if}

	<SettingsList>
		<h3 class="norm">{t("manageUsers.edit.permissions")}</h3>

		<Setting
			title={t("manageUsers.admin")}
			desc={t("manageUsers.edit.adminDesc")}
			row
		>
			<Checkbox
				name="USER_PERM_ADMIN"
				value={userHasPermission(user.permissions, UserPermission.PERM_ADMIN)}
				toggled={() => {
					userTogglePermission(UserPermission.PERM_ADMIN);
				}}
			/>
		</Setting>

		<Setting
			title={t("manageUsers.edit.requestContent")}
			desc={t("manageUsers.edit.requestContentDesc")}
			row
		>
			<Checkbox
				name="USER_PERM_REQUEST_CONTENT"
				value={userHasPermission(
					user.permissions,
					UserPermission.PERM_REQUEST_CONTENT,
				)}
				toggled={() => {
					userTogglePermission(UserPermission.PERM_REQUEST_CONTENT);
				}}
			/>
		</Setting>

		<Setting
			title={t("manageUsers.edit.autoApprove")}
			desc={t("manageUsers.edit.autoApproveDesc")}
			row
		>
			<Checkbox
				name="PERM_REQUEST_CONTENT_AUTO_APPROVE"
				value={userHasPermission(
					user.permissions,
					UserPermission.PERM_REQUEST_CONTENT_AUTO_APPROVE,
				)}
				toggled={() => {
					userTogglePermission(
						UserPermission.PERM_REQUEST_CONTENT_AUTO_APPROVE,
					);
				}}
			/>
		</Setting>

		<h3 class="norm">{t("manageUsers.edit.other")}</h3>

		<Setting
			title={t("manageUsers.edit.userType")}
			desc={t("manageUsers.edit.userTypeDesc")}
		>
			{#if !user.type || user.type === UserType.Proxy}
				<DropDown
					placeholder={t("common.unknown")}
					bind:active={user.type}
					options={[
						{
							id: 0,
							value: "Watcharr",
						},
						// {
						//   id: UserType.Jellyfin,
						//   value: "Jellyfin",
						// },
						// {
						//   id: UserType.Plex,
						//   value: "Plex",
						// },
						{
							id: UserType.Proxy,
							value: t("manageUsers.edit.proxy"),
						},
					]}
					isDropDownItem={true}
				/>
			{:else}
				<p>{t("manageUsers.edit.notSupported")}</p>
			{/if}
		</Setting>

		<div class="btns">
			<button onclick={() => save()}>{t("common.save")}</button>
		</div>
	</SettingsList>
</Modal>

<style lang="scss">
	.btns {
		display: flex;
		flex-flow: row;
		gap: 10px;

		:first-child {
			margin-left: auto;
		}

		button {
			width: max-content;
			padding-left: 15px;
			padding-right: 15px;
		}
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
</style>
