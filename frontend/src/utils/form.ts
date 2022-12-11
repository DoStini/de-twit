export interface FormValues {
    [key: string]: any
}

export const serializeForm = (formData: FormData) => {
    let obj : FormValues = {};
    for (const [key, value] of formData.entries()) {
        obj[key] = value;
    }
    return obj;
}
