<script lang="ts">
	import { resolve } from "$app/paths";
</script>

<p>
	The Jellyfin Webhook support in Watcharr allows for automatic tracking of new
	media as soon as it is started or completed in Jellyfin.
</p>

<p>
	Configuration changes must be made to both Jellyfin and Watcharr to enable
	this feature. Each step in this guide will be prefaced with either <code>
		(WATCHARR)</code
	>
	or <code>(JELLYFIN)</code> to indicate in which application you are making the change.
</p>

<ol>
	<li>
		<p><b>(JELLYFIN)</b> Install the Official Webhook plugin in Jellyfin.</p>
		<p>
			Simply go to your server dashboard, find the plugins store and look for
			one called <code>Webhook</code>. Install it.
		</p>
	</li>
	<li>
		<p>
			<b>(JELLYFIN)</b> Once the Webhook plugin is installed, you should see a
			<code>settings</code> button. Click it to enter the Webhook plugin settings.
		</p>
	</li>
	<li>
		<p>
			<b>(WATCHARR)</b> Go to
			<a href={resolve("/server")}>your server configuration page</a>, find the
			Jellyfin Webhook setting and toggle it ON.
		</p>
	</li>
	<li>
		<p>
			<b>(WATCHARR)</b> You will now see a new setting popup below the one you just
			toggled ON called Jellyfin Webhook Endpoint. Copy the address to your clipboard
			for the next step.
		</p>
	</li>
	<li>
		<p>
			<b>(JELLYFIN)</b> Click <code>Add Generic Destination</code>, then scroll
			down to the settings for the destination.
		</p>
	</li>
	<li>
		<p>
			<b>(JELLYFIN)</b> Set the <code>Webhook Name</code> to
			<code>Watcharr</code> (or anything you please).
		</p>
	</li>
	<li>
		<p>
			<b>(JELLYFIN)</b> Paste the address we copied from Watcharr into the
			<code>Webhook Url</code> field.
		</p>
		<p>
			Note: The full address copied from Watcharr <i>may</i> be wrong! Depending on
			how your server is configured, you may need to replace the hostname part of
			the address to one that your Jellyfin service can use to talk to your Watcharr
			instance.
		</p>
	</li>
	<li>
		<p>
			<b>(JELLYFIN)</b> Enable the following Notification Types:
		</p>
		<ul>
			<li>
				<p>Playback Start</p>
			</li>

			<li>
				<p>Playback Stop</p>
			</li>

			<li>
				<p>User Data Saved</p>
			</li>
		</ul>
	</li>
	<li>
		<p>
			<b>(JELLYFIN)</b> Scroll to the bottom and click <code>Save</code>.
		</p>
	</li>
	<li>
		<p>
			Done! All you have to do now is go watch something on Jellyfin and see if
			it shows up in Watcharr automatically.
		</p>
	</li>
</ol>

<h2>Problems?</h2>

<p>
	If you have any problems, the best place to start is by checking your server
	logs (Watcharr and Jellyfin).
</p>

<p>
	For example, if Jellyfin server logs show 404 errors whenever it tries sending
	the webhook, it's likely you need to double check the Webhook Url in Jellyfin
	and ensure it is correct.
</p>
