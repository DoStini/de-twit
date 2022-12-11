import type PostData from "../types/PostData";
import {env} from "$env/dynamic/public";
import {addNewPost} from "../actions/posts";

export let postsSSE: EventSource | null;

const parsePost = (item: any) => {
    const { seconds, nanos } = item.last_updated;
    item.timestamp = new Date(seconds*1000 + nanos*0.000001)
    item.username = item.user

    return item;
}

export const createPost: (post: PostData) => (void) = async (post: PostData) => {
    await fetch(env.PUBLIC_URL + "post/create", {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(post)
    });
}

export const retrieveTimeline : () => (Promise<PostData[]>) = async () => {
    const data = (await fetch(env.PUBLIC_URL + "timeline")
        .then(response => response.json())
        .then(data => {
            return data;
        }).catch(error => {
            console.log(error);
        return [];
    })).map(parsePost).sort(((item1: { timestamp: number; }, item2: { timestamp: number; }) => item2.timestamp - item1.timestamp));
    // TODO: ORDER FROM BACKEND

    return data;
}

export const registerPostsUpdate = () => {
    if (postsSSE != null) {
        postsSSE.close()
        return
    }
    postsSSE = new EventSource(env.PUBLIC_URL + "timeline/stream");
    postsSSE.onmessage = (event: MessageEvent) => {
        let response = parsePost(JSON.parse(event.data));
        addNewPost(response)
    }
}
