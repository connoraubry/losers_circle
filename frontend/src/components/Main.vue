<script lang="ts">
export default{
    data() {
        return {
            count: 0,
            longestCycle: [] 
        }
    },
    created() {
        this.updateLongestCycle()
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
            this.updateLongestCycle()
        },
        updateLongestCycle(){
            this.$http.get("http://127.0.0.1:3333/cycle")
                .then((resp) => {
                    this.longestCycle = resp.data
                })    
        }
    }
}
</script>

<template>
    <p> {{ longestCycle.length }}</p>
    <p>{{ composedOrder }}</p>
</template>
<syle>

</syle>