import {writable} from "svelte/store";
import type NotificationData from "../types/NotificationData";
import type NotificationRecord from "../types/NotificationRecord";

export const notificationsStore = writable<NotificationRecord>({})

const timeouts = []

export const launchNotification = (text: string, type: string, timeout: number) => {
    let id: string;
    notificationsStore.update((prev) => {
        id = Object.keys(prev).length.toString()
        prev[id] = {
            text: text,
            type: type,
            open: true,
            timestamp: Date.now()
        };
        return prev
    });

    timeouts.push(setTimeout(() => {
        notificationsStore.update((prev) => {
            prev[id].open = false;
            console.log(id)
            return prev
        })
    }, timeout));
}

export const closeNotification = (id: string) => {
    notificationsStore.update((prev) => {
        prev[id].open = false;
        console.log(id)
        return prev
    })
}
