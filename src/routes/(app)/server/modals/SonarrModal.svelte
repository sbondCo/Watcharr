<script lang="ts">
  import DropDown from "@/lib/DropDown.svelte";
  import Modal from "@/lib/Modal.svelte";
  import type { DropDownItem, QualityProfile, ServerConfig } from "@/types";
  import axios from "axios";

  export let serverConfig: ServerConfig;
  export let onUpdate: <K extends keyof ServerConfig>(
    name: K,
    value: ServerConfig[K],
    done?: () => void
  ) => void;
  export let onClose: () => void;

  let soDisabled = false;
  let soaDisabled = false;
  let qpDisabled = false;

  let qualityProfiles: DropDownItem[] = [];

  function testIfHostAndKeySet() {
    if (serverConfig.SONARR_HOST && serverConfig.SONARR_KEY) {
      console.log("host and key set.. loading settings data from sonarr");
      getSettingsData();
    } else {
      qualityProfiles = [];
    }
  }

  async function getSettingsData() {
    try {
      const qualityProfilesRes = await axios.get<QualityProfile[]>("/arr/son/quality_profiles");
      qualityProfiles = qualityProfilesRes.data.map((d) => {
        return { id: d.id, value: d.name };
      });
    } catch (err) {
      console.error("getSettingsData failed!", err);
      // TODO display err in modal
    }
  }

  testIfHostAndKeySet();
</script>

<Modal title="Sonarr" desc="Setup a connection to your Sonarr server" {onClose}>
  <div class="settings modal">
    <div>
      <h4 class="norm">Sonarr Host</h4>
      <h5 class="norm">Point to your Sonarr server to enable tv show requesting.</h5>
      <input
        type="text"
        placeholder="https://sonarr.example.com"
        bind:value={serverConfig.SONARR_HOST}
        on:blur={() => {
          soDisabled = true;
          onUpdate("SONARR_HOST", serverConfig.SONARR_HOST, () => {
            soDisabled = false;
            testIfHostAndKeySet();
          });
        }}
        disabled={soDisabled}
      />
    </div>
    <div>
      <h4 class="norm">Sonarr Key</h4>
      <h5 class="norm">API key for your Sonarr instance.</h5>
      <input
        type="text"
        placeholder="dGhhbmtzIGZvciB1c2luZyB3YXRjaGFyciA6KQ=="
        bind:value={serverConfig.SONARR_KEY}
        on:blur={() => {
          soaDisabled = true;
          onUpdate("SONARR_KEY", serverConfig.SONARR_KEY, () => {
            soaDisabled = false;
            testIfHostAndKeySet();
          });
        }}
        disabled={soaDisabled}
      />
    </div>
    <div>
      <h4 class="norm">Quality Profile</h4>
      <h5 class="norm">Default quality profile when adding shows.</h5>
      <DropDown
        placeholder={qualityProfiles?.length == 0
          ? "Add a Host and Key to view"
          : "Select a quality profile"}
        options={qualityProfiles}
        bind:active={serverConfig.SONARR_QUALITY_PROFILE}
        onChange={() => {
          qpDisabled = true;
          onUpdate("SONARR_QUALITY_PROFILE", serverConfig.SONARR_QUALITY_PROFILE, () => {
            qpDisabled = false;
          });
        }}
        disabled={qpDisabled}
      />
    </div>
  </div>
</Modal>

<style lang="scss">
  // Settings in modals
  .modal {
    margin-top: 15px;
  }
</style>
