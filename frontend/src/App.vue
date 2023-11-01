<template>
  <v-app>
    <v-main>
      <Header />
      <Main />
      <Footer />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
  import Header from '@/components/Header.vue'
  import Footer from './components/Footer.vue'
  import Main from './components/Main.vue'
</script>
