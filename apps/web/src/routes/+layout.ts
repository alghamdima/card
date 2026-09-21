// The app is a static single-page app served by nginx: everything renders in the browser,
// which also lets modules use localStorage/window without SSR guards.
export const ssr = false;
export const prerender = false;
