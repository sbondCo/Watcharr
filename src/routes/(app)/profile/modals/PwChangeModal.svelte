<script lang="ts">
  import Form from "@/lib/forms/Form.svelte";
  import Modal from "@/lib/Modal.svelte";
  import type { ChangePasswordForm } from "@/types";

  export let userName: string | undefined;
  export let onClose: () => void;
  export let changepswd: ChangePasswordForm = {
    username: "",
    currentPassword: "",
    newPassword: "",
    reEnteredNewPassword: ""
  };

  let error: string;
  let errs: string[] = [];

  function checkForm() {
    errs = [];
    console.log("Status: Checking if any form inputs are missing");
    if (!changepswd.currentPassword) {
      errs.push("Current Password");
    }
    if (!changepswd.newPassword) {
      errs.push("New Password");
    }
    if (!changepswd.reEnteredNewPassword) {
      errs.push("Re-Entered New Password");
    }
    if (errs.length > 0) {
      error = `Missing required params: ${errs.join(", ")}`;
      console.log(`Error: following form inputs are missing:\n${errs.join("\n")}`);
    } else {
      console.log("Status: All form inputs are present")
      checkFormPasswordsMatch()
    }
  }

  function checkFormPasswordsMatch() {
    console.log("Status: Checking if new password and re-entered new password match");
    if (changepswd.newPassword !== changepswd.reEnteredNewPassword) {
      error = "New password and re-entered new password do not match";
      console.log("Error: New password and re-entered new password do not match");
    } else {
      console.log("Status: New password and re-entered new password match");
      error = "";
    }
  }

  function handleSubmit(ev: SubmitEvent) {
    checkForm();
    // checkFormPasswordsMatch();
    if (!error) {
      console.log("Form inputs are valid");
      const fd = new FormData(ev.target! as HTMLFormElement);
      changepswd.username = fd.get("username") as string;
      changepswd.currentPassword = fd.get("current-password") as string;
      changepswd.newPassword = fd.get("new-password") as string;
      changepswd.reEnteredNewPassword = fd.get("re-entered-new-password") as string;
      console.log(changepswd);
    }
  }
</script>

<Modal
  title="Change Password"
  desc="Use the below form to change your password for account {userName}"
  {onClose}
>
  {#if error}
    <span class="error">{error}!</span>
  {/if}
  <form on:submit|preventDefault={handleSubmit}>
    <div class="form-input-container">
      <div class="form-input">
        <!--Hiding username info as it is still useful to password managers-->
        <!--https://www.chromium.org/developers/design-documents/create-amazing-password-forms/#use-hidden-fields-for-implicit-information-->
        <label for="username" id="username-label">Username</label>
        <input
          type="text"
          name="username"
          autocomplete="username"
          id="username-input"
          value={userName}
        />
      </div>
      <div class="form-input">
        <label for="current-password">Current Password</label>
        <input
          type="password"
          name="current-password"
          placeholder="Current password"
          autocomplete="current-password"
          bind:value={changepswd.currentPassword}
        />
      </div>
      <div class="form-input">
        <label for="new-password">New Password</label>
        <input
          type="password"
          name="new-password"
          placeholder="New password"
          autocomplete="new-password"
          bind:value={changepswd.newPassword}
        />
      </div>
      <div class="form-input">
        <label for="re-entered-new-password">Re-enter New Password</label>
        <input
          type="password"
          name="re-entered-new-password"
          placeholder="Re-enter new password"
          autocomplete="new-password"
          bind:value={changepswd.reEnteredNewPassword}
        />
      </div>
      <button type="submit">Change Password</button>
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
