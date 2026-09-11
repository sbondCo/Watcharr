<script lang="ts">
	interface Props {
		src: string;
		alt: string;
		maxWidth?: string;
	}

	let { src, alt, maxWidth }: Props = $props();
</script>

<img {src} {alt} style:max-width={maxWidth} loading="lazy" />
