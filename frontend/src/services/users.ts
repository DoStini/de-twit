import {env} from "$env/dynamic/public";
import type UserData from "../types/UserData";

export const searchUser : (username: string) => (Promise<UserData>) = async (username) => {
    const data = await fetch(env.PUBLIC_URL + username)
        .then(async response => {
            const json = await response.json()
            return {response, json}
        })
        .then(({response, json}) => {
            if (!response.ok) {
                throw json
            }
            return json
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
