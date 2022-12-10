<script lang="ts">
    import "../app.postcss";
    import Posts from "../components/post/Posts.svelte";
    import Fab from "../components/Fab.svelte";
    import { addNewPost, postsStore } from "../actions/posts.js"
    import type PostData from "../types/PostData.js";
    import {newPostsStore} from "../actions/posts";
    import {onMount} from "svelte";
    import {retrieveTimeline} from "../services/posts";
    import ErrorAlert from "../components/alerts/ErrorAlert.svelte";

    let posts: PostData[]
    postsStore.subscribe((value) => posts = value)

    let newPosts: PostData[]
    newPostsStore.subscribe((value) => newPosts = value)

    onMount(async () => {
        try {
            const posts = await retrieveTimeline()
            postsStore.set(posts)
        } catch (e) {

        }
    })
</script>

<Posts posts={posts} newPosts={newPosts}/>

<Fab action={() => addNewPost({
    text: "Ola mundo",
    username: "andremoreira9",
    timestamp: new Date(Date.now())
})}/>
