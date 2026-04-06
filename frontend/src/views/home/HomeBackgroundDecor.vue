<template>
  <div class="home-background-decor pointer-events-none absolute inset-0 overflow-hidden">
    <div class="home-background-decor__orb home-background-decor__orb--top-right"></div>
    <div class="home-background-decor__orb home-background-decor__orb--bottom-left"></div>
    <div class="home-background-decor__orb home-background-decor__orb--center"></div>
    <div class="home-background-decor__orb home-background-decor__orb--lower-right"></div>
    <div class="home-background-decor__grid absolute inset-0"></div>
  </div>
</template>

<style scoped>
.home-background-decor__orb {
  position: absolute;
  border-radius: 999px;
  filter: blur(var(--theme-home-decor-orb-blur));
}

.home-background-decor__orb--top-right {
  top: var(--theme-home-decor-orb-top-right-offset);
  right: var(--theme-home-decor-orb-top-right-offset);
  width: var(--theme-home-decor-orb-large-size);
  height: var(--theme-home-decor-orb-large-size);
}

.home-background-decor__orb--top-right {
  background: color-mix(in srgb, var(--theme-accent) 20%, transparent);
}

.home-background-decor__orb--bottom-left {
  bottom: var(--theme-home-decor-orb-bottom-left-offset);
  left: var(--theme-home-decor-orb-bottom-left-offset);
  width: var(--theme-home-decor-orb-large-size);
  height: var(--theme-home-decor-orb-large-size);
}

.home-background-decor__orb--bottom-left {
  background: color-mix(in srgb, var(--theme-accent-strong) 15%, transparent);
}

.home-background-decor__orb--center {
  left: var(--theme-home-decor-orb-center-left);
  top: var(--theme-home-decor-orb-center-top);
  width: var(--theme-home-decor-orb-medium-size);
  height: var(--theme-home-decor-orb-medium-size);
}

.home-background-decor__orb--center {
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, transparent);
}

.home-background-decor__orb--lower-right {
  right: var(--theme-home-decor-orb-lower-right-offset);
  bottom: var(--theme-home-decor-orb-lower-right-offset);
  width: var(--theme-home-decor-orb-small-size);
  height: var(--theme-home-decor-orb-small-size);
}

.home-background-decor__orb--lower-right {
  background: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 10%, transparent);
}

.home-background-decor__grid {
  background-image:
    linear-gradient(
      color-mix(in srgb, var(--theme-accent-soft) 22%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      90deg,
      color-mix(in srgb, var(--theme-accent-soft) 22%, transparent) 1px,
      transparent 1px
    );
  background-size: var(--theme-home-decor-grid-size) var(--theme-home-decor-grid-size);
}
</style>
