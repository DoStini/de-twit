const weekday = ["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"];

const twoDigitNumber = (number: number) => ("0" + number).slice(-2);

export const dateToString = (date: Date) : String => {
    return `${weekday[date.getDay()]} ${date.getDate()}-${date.getMonth() + 1}-${date.getFullYear()} ${twoDigitNumber(date.getHours())}:${twoDigitNumber(date.getMinutes())}:${twoDigitNumber(date.getSeconds())}`
}