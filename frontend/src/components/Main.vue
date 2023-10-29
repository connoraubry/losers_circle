<script lang="ts">
export default{
    data() {
        return {
            count: 0,
            longestCycle: [] 
        }
    },
    computed: {
        composedOrder(){
            return this.longestCycle.join(" > ")
        }
    },
    methods: {
        clickButtonHandler() {
            console.log("Click!")
            this.count ++
            this.$http.get("http://127.0.0.1:3333/cycle")
                .then((resp) => {
                    this.longestCycle = resp.data
                })     
        }
    }
}
</script>

<template>
    <button @click="clickButtonHandler">Click me</button>
    <p>{{ composedOrder }}</p>
    <p>{{ count }}</p>

</template>
<syle>

</syle>