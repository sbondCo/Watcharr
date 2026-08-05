<script lang="ts">
	import PersonDetailsPage from "@/lib/content/PersonDetailsPage.svelte";
	import { MediaTypeE } from "@/types.js";

	let { data } = $props();
</script>

<PersonDetailsPage
	personId={data.authorId!}
	apiPath="/book/author"
	mediaType={MediaTypeE.olBookAuthor}
/>
