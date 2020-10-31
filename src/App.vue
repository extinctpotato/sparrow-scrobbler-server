<template>
  <div id="app">
    <b-navbar toggleable="lg" type="dark" variant="dark">
      <b-navbar-brand href="#">NavBar</b-navbar-brand>
    </b-navbar>
  </div>
</template>

<script>
import { BNavbar } from 'bootstrap-vue'

export default {
  name: 'App',
  components: {
    'b-navbar': BNavbar
  }
}
</script>

<style>
</style>
