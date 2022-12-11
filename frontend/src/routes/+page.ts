import {retrieveTimeline} from "../services/posts";
import {postsStore} from "../actions/posts";

export async function load() {
    try {
        const posts = await retrieveTimeline()
        postsStore.set(posts)
        return {}
    } catch (e) {
        console.error(e)
        postsStore.set([])
        return {}
    }
}