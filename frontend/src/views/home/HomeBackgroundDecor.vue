<template>
  <div class="home-bg pointer-events-none absolute inset-0 overflow-hidden">
    <!-- Ambient orbs: shared across themes, differently tinted per theme. -->
    <div class="home-bg__orb home-bg__orb--top-right"></div>
    <div class="home-bg__orb home-bg__orb--bottom-left"></div>
    <div class="home-bg__orb home-bg__orb--center"></div>
    <div class="home-bg__orb home-bg__orb--lower-right"></div>

    <!-- Factory: blueprint grid + scanlines + corner rivets; Claude: paper noise + ornamental seal. -->
    <div class="home-bg__grid absolute inset-0"></div>
    <div class="home-bg__scanlines absolute inset-0"></div>
    <div class="home-bg__paper absolute inset-0"></div>
    <div class="home-bg__rivet home-bg__rivet--tl"></div>
    <div class="home-bg__rivet home-bg__rivet--tr"></div>
    <div class="home-bg__rivet home-bg__rivet--bl"></div>
    <div class="home-bg__rivet home-bg__rivet--br"></div>
  </div>
</template>

<style scoped>
.home-bg__orb {
  position: absolute;
  border-radius: 999px;
  filter: blur(var(--theme-home-decor-orb-blur));
}

.home-bg__orb--top-right {
  top: var(--theme-home-decor-orb-top-right-offset);
  right: var(--theme-home-decor-orb-top-right-offset);
  width: var(--theme-home-decor-orb-large-size);
  height: var(--theme-home-decor-orb-large-size);
  background: color-mix(in srgb, var(--theme-accent) 20%, transparent);
}

.home-bg__orb--bottom-left {
  bottom: var(--theme-home-decor-orb-bottom-left-offset);
  left: var(--theme-home-decor-orb-bottom-left-offset);
  width: var(--theme-home-decor-orb-large-size);
  height: var(--theme-home-decor-orb-large-size);
  background: color-mix(in srgb, var(--theme-accent-strong) 15%, transparent);
}

.home-bg__orb--center {
  left: var(--theme-home-decor-orb-center-left);
  top: var(--theme-home-decor-orb-center-top);
  width: var(--theme-home-decor-orb-medium-size);
  height: var(--theme-home-decor-orb-medium-size);
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, transparent);
}

.home-bg__orb--lower-right {
  right: var(--theme-home-decor-orb-lower-right-offset);
  bottom: var(--theme-home-decor-orb-lower-right-offset);
  width: var(--theme-home-decor-orb-small-size);
  height: var(--theme-home-decor-orb-small-size);
  background: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 10%, transparent);
}

.home-bg__grid {
  display: var(--theme-home-bg-grid-display);
  background-image: var(--theme-home-bg-grid-image);
  background-size: var(--theme-home-decor-grid-size) var(--theme-home-decor-grid-size);
  mask-image: var(--theme-home-bg-grid-mask);
  -webkit-mask-image: var(--theme-home-bg-grid-mask);
}

.home-bg__scanlines {
  display: var(--theme-home-bg-scanlines-display);
  background-image: var(--theme-home-bg-scanlines-image);
  background-size: var(--theme-home-bg-scanlines-size);
  opacity: var(--theme-home-bg-scanlines-opacity);
  mix-blend-mode: var(--theme-home-bg-scanlines-blend);
}

.home-bg__paper {
  display: var(--theme-home-bg-paper-display);
  background-image: var(--theme-home-bg-paper-image);
  background-size: var(--theme-home-bg-paper-size);
  opacity: var(--theme-home-bg-paper-opacity);
  mix-blend-mode: var(--theme-home-bg-paper-blend);
}

.home-bg__rivet {
  display: var(--theme-home-bg-rivet-display);
  position: absolute;
  width: var(--theme-home-bg-rivet-size);
  height: var(--theme-home-bg-rivet-size);
  background: var(--theme-home-bg-rivet-color);
  border-radius: var(--theme-home-bg-rivet-radius);
}

.home-bg__rivet--tl { top: 16px; left: 16px; }
.home-bg__rivet--tr { top: 16px; right: 16px; }
.home-bg__rivet--bl { bottom: 16px; left: 16px; }
.home-bg__rivet--br { bottom: 16px; right: 16px; }
</style>
