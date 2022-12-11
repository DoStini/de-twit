import {env} from "$env/dynamic/public";
import type UserData from "../types/UserData";

export const searchUser : (username: string) => (Promise<UserData>) = async (username) => {
    const data = await fetch(env.PUBLIC_URL + username)
        .then(response => response.json())
        .then(data => {
            data.posts = data.posts
            return data;
        });

    return data;
}

export const followUser : (username: string, shouldFollow: boolean) => (Promise<void>) = async (username: string, shouldFollow: boolean) => {
    const route = `${username}/${shouldFollow ? '' : 'un'}follow`
    await fetch(env.PUBLIC_URL + route, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
        },
    });
}
