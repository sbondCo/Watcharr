<span class="loader">
	<svg viewBox="0 0 50 50">
		<circle
			cx="25"
			cy="25"
			r="20"
			fill="none"
			stroke="currentColor"
			stroke-width="6"
			stroke-linecap="round"
		/>
	</svg>
</span>

<style lang="scss">
	.loader {
		$animDur: 1.4s;
		display: flex;
		width: 16px;
		height: 16px;
		animation: rotation $animDur linear infinite;

		svg {
			circle {
				stroke-dasharray: 80px 200px;
				stroke-dashoffset: 0px;
				animation: circle-size $animDur ease-in-out infinite;
			}
		}
	}

	@keyframes rotation {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}

	@keyframes circle-size {
		0% {
			stroke-dasharray: 1px 200px;
			stroke-dashoffset: 0px;
		}
		50% {
			stroke-dasharray: 100px 200px;
			stroke-dashoffset: -15px;
		}
		100% {
			stroke-dasharray: 1px 200px;
			stroke-dashoffset: -124px;
		}
	}
</style>
