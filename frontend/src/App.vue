<template>
  <div id="app" class="w-full h-full">
    <router-view :key="$route.fullPath" />
  </div>
</template>

<script>
export default {
  name: "App",
};
</script>