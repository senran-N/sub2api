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

/* Default: hidden. Each theme opts in below. */
.home-bg__grid,
.home-bg__scanlines,
.home-bg__paper,
.home-bg__rivet {
  display: none;
}

/* ============== Factory: blueprint grid + scanlines + rivets ============== */
:root[data-brand-theme='factory'] .home-bg__grid {
  display: block;
  background-image:
    linear-gradient(
      color-mix(in srgb, var(--theme-page-text) 10%, transparent) 1px,
      transparent 1px
    ),
    linear-gradient(
      90deg,
      color-mix(in srgb, var(--theme-page-text) 10%, transparent) 1px,
      transparent 1px
    );
  background-size: var(--theme-home-decor-grid-size) var(--theme-home-decor-grid-size);
  mask-image: radial-gradient(ellipse at center, black 40%, transparent 85%);
  -webkit-mask-image: radial-gradient(ellipse at center, black 40%, transparent 85%);
}

:root[data-brand-theme='factory'] .home-bg__scanlines {
  display: block;
  background: repeating-linear-gradient(
    0deg,
    transparent 0,
    transparent 3px,
    color-mix(in srgb, var(--theme-page-text) 4%, transparent) 3px,
    color-mix(in srgb, var(--theme-page-text) 4%, transparent) 4px
  );
  opacity: 0.5;
}

:root[data-brand-theme='factory'] .home-bg__rivet {
  display: block;
  position: absolute;
  width: 10px;
  height: 10px;
  background: var(--theme-page-text);
  border-radius: 0;
}

:root[data-brand-theme='factory'] .home-bg__rivet--tl { top: 16px; left: 16px; }
:root[data-brand-theme='factory'] .home-bg__rivet--tr { top: 16px; right: 16px; }
:root[data-brand-theme='factory'] .home-bg__rivet--bl { bottom: 16px; left: 16px; }
:root[data-brand-theme='factory'] .home-bg__rivet--br { bottom: 16px; right: 16px; }

.dark[data-brand-theme='factory'] .home-bg__scanlines {
  opacity: 0.3;
}

/* ============== Claude: warm paper wash + soft radial glow ============== */
:root[data-brand-theme='claude'] .home-bg__paper {
  display: block;
  background-image:
    radial-gradient(
      circle at 20% 20%,
      color-mix(in srgb, var(--theme-accent) 14%, transparent),
      transparent 45%
    ),
    radial-gradient(
      circle at 80% 85%,
      color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, transparent),
      transparent 50%
    );
}

:root[data-brand-theme='claude'] .home-bg__scanlines {
  display: block;
  background-image:
    radial-gradient(
      color-mix(in srgb, var(--theme-page-text) 4%, transparent) 1px,
      transparent 1px
    );
  background-size: 6px 6px;
  opacity: 0.45;
  mix-blend-mode: multiply;
}

.dark[data-brand-theme='claude'] .home-bg__scanlines {
  mix-blend-mode: screen;
  opacity: 0.2;
}
</style>
