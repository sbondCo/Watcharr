<div class="container">
  <div class="spinner" />
</div>

<style lang="scss">
  .container {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .spinner {
    width: 80px;
    aspect-ratio: 1;
    border: 10px solid $bg-color;
    background: radial-gradient(farthest-side, $text-color 98%, $bg-color) 0 0/20px 20px,
      radial-gradient(farthest-side, $text-color 98%, $bg-color) 100% 0/20px 20px,
      radial-gradient(farthest-side, $text-color 98%, $bg-color) 100% 100%/20px 20px,
      radial-gradient(farthest-side, $text-color 98%, $bg-color) 0 100%/20px 20px,
      linear-gradient($text-color 0 0) center/40px 40px, $bg-color;
    background-repeat: no-repeat;
    filter: blur(4px) contrast(10);
    animation: spin 800ms infinite;
  }

  @keyframes spin {
    100% {
      background-position: 100% 0, 100% 100%, 0 100%, 0 0, center;
    }
  }
</style>
