<script lang="ts">
    import Fab from "../components/Fab.svelte";
    import {addNewPost, postsStore, refreshTimeline} from "../actions/posts.js"
    import type PostData from "../types/PostData.js";
    import {newPostsStore} from "../actions/posts";
    import {onMount} from "svelte";
    import {registerPostsUpdate, retrieveTimeline} from "../services/posts";
    import NewPostModal from "../components/post/NewPostModal.svelte";
    import {closeNewPostModal, openNewPostModal} from "../actions/newPostModal.js";
    import {createPost} from "../services/posts.js";
    import {env} from "$env/dynamic/public";
    import Posts from "../components/post/Posts.svelte";

    let posts: PostData[]
    postsStore.subscribe((value) => posts = value)

    let newPosts: PostData[]
    newPostsStore.subscribe((value) => newPosts = value)

    let loading: boolean = false

    let handleCreatePost = async (data) => {
        loading = true
        try {
            const post = {
                username: env.PUBLIC_USERNAME,
                text: data.content,
                timestamp: new Date()
            };

            await createPost(post);
            addNewPost(post)

            closeNewPostModal()
            loading = false
            return true
        } catch (e) {
            console.error(e)
            loading = false
            return true
        }
    }

    onMount(() => {
        const sse = registerPostsUpdate(addNewPost)
        return () => sse.close()
    })

</script>

<Posts emptyPostsMessage="Nothing to show! Post something or find other users to follow!"
       mainTimeline={true}
       newPosts={newPosts}
       posts={posts}
       refreshTimeline={refreshTimeline}/>

<Fab action={openNewPostModal}/>

<NewPostModal loading={loading} close={closeNewPostModal} submit={handleCreatePost}/>
