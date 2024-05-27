// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import { decodeMessage } from './decoder';

export const convertDateToSeconds = (days, hours, minutes, seconds) => {
    let totalSeconds = 0;
    totalSeconds += days !== 0 ? days * 86400 : 0;
    totalSeconds += hours !== 0 ? hours * 3600 : 0;
    totalSeconds += minutes !== 0 ? minutes * 60 : 0;
    totalSeconds += seconds !== 0 ? seconds : 0;
    return totalSeconds;
};

export const convertSecondsToDateObject = (seconds) => {
    const days = Math.floor(seconds / 86400);
    seconds -= days * 86400;
    const hours = Math.floor(seconds / 3600);
    seconds -= hours * 3600;
    const minutes = Math.floor(seconds / 60);
    seconds -= minutes * 60;

    return {
        days,
        hours,
        minutes,
        seconds
    };
};

export const convertSecondsToDate = (seconds, short = false) => {
    const days = Math.floor(seconds / 86400);
    seconds -= days * 86400;
    const hours = Math.floor(seconds / 3600);
    seconds -= hours * 3600;
    const minutes = Math.floor(seconds / 60);
    seconds -= minutes * 60;

    let result = '';
    if (days > 0) {
        result = days === 1 ? '1 day' : `${days} days`;
        if (hours > 0) {
            result = hours === 1 ? `${result}, 1 hour` : `${result}, ${hours} hours`;
        }
        if (minutes > 0) {
            result = minutes === 1 ? `${result}, 1 minute` : `${result}, ${minutes} minutes`;
        }
        if (seconds > 0) {
            result = seconds === 1 ? `${result}, 1 second` : `${result}, ${seconds} seconds`;
        }
    } else if (hours > 0) {
        result = hours === 1 ? '1 hour' : `${hours} hours`;
        if (minutes > 0) {
            result = minutes === 1 ? `${result}, 1 minute` : `${result}, ${minutes} minutes`;
        }
        if (seconds > 0) {
            result = seconds === 1 ? `${result}, 1 second` : `${result}, ${seconds} seconds`;
        }
    } else if (minutes > 0) {
        result = minutes === 1 ? '1 minute' : `${minutes} minutes`;
        if (seconds > 0) {
            result = seconds === 1 ? `${result}, 1 second` : `${result}, ${seconds} seconds`;
        }
    } else if (seconds > 0) {
        result = seconds === 1 ? '1 second' : `${seconds} seconds`;
    }

    const spliter = result.split(',');

    for (let i = 0; i < spliter.length; i++) {
        if (i === 0) {
            result = spliter[0];
        } else if (i < spliter.length - 1) {
            result = `${result}, ${spliter[i]}`;
        } else {
            result = `${result} and ${spliter[i]}`;
        }
    }
    if (short) {
        const replacements = [
            { search: /\bdays?\b/g, replace: 'd' },
            { search: /\bhours?\b/g, replace: 'h' },
            { search: /\bminutes?\b/g, replace: 'm' },
            { search: /\bseconds?\b/g, replace: 's' }
        ];
        let outputString = result;
        for (const replacement of replacements) {
            outputString = outputString.replace(replacement.search, replacement.replace);
        }
        result = outputString;
    }

    return result;
};

export const parsingDate = (date, withSeconds = false, withTime = true) => {
    if (date) {
        var second = withSeconds ? 'numeric' : undefined;
        var time = withTime ? 'numeric' : undefined;

        var options = { year: 'numeric', month: 'short', day: 'numeric', hour: time, minute: time, second: second };
        return new Date(date).toLocaleDateString([], options);
    } else {
        return '';
    }
};

export const parsingDateWithotTime = (date) => {
    if (date) {
        var options = { year: 'numeric', month: 'short', day: 'numeric' };
        return new Date(date).toLocaleDateString([], options);
    } else return '';
};

function isFloat(n) {
    return Number(n) === n && n % 1 !== 0;
}
export const convertBytesToGb = (bytes) => {
    return bytes / 1024 / 1024 / 1024;
};

export const convertBytes = (bytes, round) => {
    const KB = 1024;
    const MB = KB * 1024;
    const GB = MB * 1024;
    const TB = GB * 1024;
    const PB = TB * 1024;

    if (bytes < KB && bytes > 0) {
        return `${bytes} Bytes`;
    } else if (bytes >= KB && bytes < MB) {
        const parsing = isFloat(bytes / KB) ? Math.round((bytes / KB + Number.EPSILON) * 100) / 100 : bytes / KB;
        return `${round ? Math.trunc(parsing) : parsing} KB`;
    } else if (bytes >= MB && bytes < GB) {
        const parsing = isFloat(bytes / MB) ? Math.round((bytes / MB + Number.EPSILON) * 100) / 100 : bytes / MB;
        return `${round ? Math.trunc(parsing) : parsing} MB`;
    } else if (bytes >= GB && bytes < TB) {
        const parsing = isFloat(bytes / GB) ? Math.round((bytes / GB + Number.EPSILON) * 100) / 100 : bytes / GB;
        return `${round ? Math.trunc(parsing) : parsing} GB`;
    } else if (bytes >= TB && bytes < PB) {
        const parsing = isFloat(bytes / TB) ? Math.round((bytes / TB + Number.EPSILON) * 100) / 100 : bytes / TB;
        return `${round ? Math.trunc(parsing) : parsing} TB`;
    } else if (bytes >= PB) {
        const parsing = isFloat(bytes / PB) ? Math.round((bytes / PB + Number.EPSILON) * 100) / 100 : bytes / PB;
        return `${round ? Math.trunc(parsing) : parsing} PB`;
    } else {
        return '0 Bytes';
    }
};

export const capitalizeFirst = (str) => {
    return str?.charAt(0).toUpperCase() + str?.slice(1).toLowerCase();
};

export const filterArray = (arr1, arr2) => {
    const filtered = arr1.filter((el) => {
        return arr2.indexOf(el.name) === -1;
    });
    return filtered;
};

export const stationFilterArray = (arr1, arr2) => {
    const filtered = arr1.filter((station) => {
        return arr2.indexOf(station.station.name) === -1;
    });
    return filtered;
};

export const isThereDiff = (s1, s2) => {
    if (s1 === s2) {
        return false;
    }
    return true;
};

export const getUnique = (obj) => {
    const uniqueIds = [];

    const unique = obj?.filter((element) => {
        const isDuplicate = uniqueIds?.includes(element.name);

        if (!isDuplicate) {
            uniqueIds.push(element.name);

            return true;
        }

        return false;
    });
    return unique;
};

export const diffDate = (date) => {
    var msDiff = new Date(date).getTime() - new Date().getTime(); //Future date - current date
    var dayDiff = Math.floor(msDiff / (1000 * 60 * 60 * 24)) * -1;
    if (dayDiff === 1) {
        return 'Today';
    }
    return `${dayDiff} days ago`;
};

export const hex_to_ascii = (input) => {
    if (typeof input === 'string' && /^[0-9a-fA-F]+$/.test(input)) {
        let str = '';
        try {
            str = decodeURIComponent(input.replace(/[0-9a-f]{2}/g, '%$&'));
        } catch {
            return input;
        }
        return str;
    } else if (typeof input === 'number') {
        return String.fromCharCode(input);
    } else if (typeof input === 'string') {
        return input;
    } else {
        return input;
    }
};

export const ascii_to_hex = (str) => {
    let hex = '';
    for (let i = 0; i < str.length; i++) {
        hex += str.charCodeAt(i).toString(16);
    }
    return hex;
};

export const isHexString = (str) => {
    const hexChars = /^[0-9a-fA-F]+$/;
    if (hexChars.test(str)) {
        return true;
    }

    return false;
};

export const compareObjects = (object1, object2) => {
    const keys1 = Object.keys(object1);
    const keys2 = Object.keys(object2);
    if (keys1.length !== keys2.length) {
        return false;
    }
    for (let key of keys1) {
        if (object1[key] !== object2[key]) {
            return false;
        }
    }
    return true;
};

export const msToUnits = (value) => {
    const second = 1000;
    const minute = second * 60;
    const hour = minute * 60;
    const day = hour * 24;
    let parsing = 0;
    switch (true) {
        case value < second && value >= 100:
            return `${value?.toLocaleString()} ms`;
        case value >= second && value < minute:
            parsing = isFloat(value / second) ? Math.round((value / second + Number.EPSILON) * 100) / 100 : value / second;
            if (parsing === 1) {
                return `${parsing} second`;
            } else {
                return `${parsing?.toLocaleString()} seconds`;
            }
        case value >= minute && value < hour:
            parsing = isFloat(value / minute) ? Math.round((value / minute + Number.EPSILON) * 100) / 100 : value / minute;
            if (parsing === 1) {
                return `${parsing} minute`;
            } else {
                return `${parsing?.toLocaleString()} minutes`;
            }
        case value >= hour && value < day:
            parsing = isFloat(value / hour) ? Math.round((value / hour + Number.EPSILON) * 100) / 100 : value / hour;
            if (parsing === 1) {
                return `${parsing} hour`;
            } else {
                return `${parsing?.toLocaleString()} hours`;
            }
        case value >= day:
            parsing = isFloat(value / day) ? Math.round((value / day + Number.EPSILON) * 100) / 100 : value / day;
            if (parsing === 1) {
                return `${parsing} day`;
            } else {
                return `${parsing?.toLocaleString()} days`;
            }
        default:
            break;
    }
};

export const generateName = (value) => {
    return value?.trimStart().replaceAll(' ', '-')?.toLowerCase();
};

export const idempotencyValidator = (value, idempotencyType) => {
    const idempotencyOptions = ['Milliseconds', 'Seconds', 'Minutes', 'Hours'];

    return new Promise((resolve, reject) => {
        if (value !== '') {
            switch (idempotencyType) {
                case idempotencyOptions[0]:
                    if (value < 100) {
                        return reject('Has to be greater than 100ms');
                    }
                    if (value > 8.64e7) {
                        return reject('Has to be lower than 24 hours');
                    } else {
                        return resolve();
                    }
                case idempotencyOptions[1]:
                    if (value > 86400) {
                        return reject('Has to be lower than 24 hours');
                    } else {
                        return resolve();
                    }
                case idempotencyOptions[2]:
                    if (value > 1440) {
                        return reject('Has to be lower than 24 hours');
                    } else {
                        return resolve();
                    }
                case idempotencyOptions[3]:
                    if (value > 24) {
                        return reject('Has to be lower than 24 hours');
                    } else {
                        return resolve();
                    }
                default:
                    break;
            }
        } else {
            return reject('Please input idempotency value');
        }
    });
};

export const tieredStorageTimeValidator = (value) => {
    if (value === 0) {
        return 'Please input tiered storage value';
    }
    if (value < 5) {
        return 'Has to be higher than 5 seconds';
    } else if (value > 3600) {
        return 'Has to be 1 hour or lower';
    } else {
        return '';
    }
};

export const partitionsValidator = (value) => {
    if (value <= 0) {
        return 'At least 1 partition is required';
    }
    if (value > 10000) {
        return 'Max number of partitions is: 10,000';
    } else {
        return '';
    }
};

export const replicasConvertor = (value, stringToNumber) => {
    if (stringToNumber) {
        switch (value) {
            case 'No HA (1)':
                return 1;
            case 'HA (3)':
                return 3;
            case 'Super HA (5)':
                return 5;
            default:
                return 1;
        }
    } else {
        if (value >= 1 && value < 3) return 'No HA (1)';
        else if (value >= 3 && value < 5) return 'HA (3)';
        else if (value >= 5) return 'Super HA (5)';
        else return 'No HA (1)';
    }
};

const isJsonString = (str) => {
    try {
        JSON.parse(str);
    } catch (e) {
        return false;
    }
    return true;
};

export const messageParser = (type, data) => {
    switch (type) {
        case 'string':
            return hex_to_ascii(data);
        case 'json':
            let str = hex_to_ascii(data);
            if (isJsonString(str)) {
                return JSON.stringify(JSON.parse(str), null, 2);
            } else {
                return str;
            }
        case 'protobuf':
            return JSON.stringify(decodeMessage(data), null, 2);
        case 'bytes':
            const isHexStr = isHexString(data);
            if (isHexStr) {
                return data;
            } else {
                return ascii_to_hex(data);
            }
        default:
            return hex_to_ascii(data);
    }
};

export const compareVersions = (a, b) => {
    const versionA = a.split('.');
    const versionB = b.split('.');

    for (let i = 0; i < versionA.length; i++) {
        const numberA = parseInt(versionA[i]);
        const numberB = parseInt(versionB[i]);
        if (numberA > numberB) {
            return true;
        } else if (numberA < numberB) {
            return false;
        }
    }
    return true;
};

export const convertArrayToObject = (array) => {
    if (array.length === 0) return {};
    const obj = {};
    for (const item of array) {
        obj[item.key] = item.value;
    }
    return obj;
};

const predefinedPairs = [
    { key: 'name', value: 'John' },
    { key: 'age', value: '25' },
    { key: 'email', value: 'john@example.com' },
    { key: 'address', value: '123 Main St' },
    { key: 'phone', value: '555-124' },
    { key: 'city', value: 'New York' },
    { key: 'country', value: 'USA' },
    { key: 'occupation', value: 'Software Engineer' },
    { key: 'interests', value: 'Sports,' },
    { key: 'hobby', value: 'Cooking' }
];

export const generateJSONWithMaxLength = (maxLength) => {
    function generateValue(currentLength) {
        if (currentLength >= maxLength) {
            return null;
        }

        const result = {};
        let remainingLength = maxLength - currentLength;

        while (remainingLength > 0) {
            const pairIndex = Math.floor(Math.random() * predefinedPairs.length);
            const pair = predefinedPairs[pairIndex];

            if (result[pair.key] === undefined) {
                result[pair.key] = pair.value;
                const pairLength = JSON.stringify({ [pair.key]: pair.value }).length;
                remainingLength -= pairLength;
            }

            if (remainingLength <= 0) {
                break;
            }
        }

        return result;
    }

    let result = generateValue(0, 3);
    result = JSON.stringify(result, null, 2);
    if (result?.length > maxLength) {
        result = generateValue(0, 3);
        result = JSON.stringify(result, null, 2);
    }
    return result;
};

export const extractValueFromURL = (url, type) => {
    const regex = /\/stations\/([^/?]+)(?:\/(\d+))?/;

    const match = url.match(regex);

    if (match && match.length >= 2) {
        const stationName = match[1];

        if (type === 'name') {
            return stationName;
        } else if (type === 'id' && match.length >= 3) {
            const stationID = match[2];
            return stationID;
        }
    }

    return null;
};

export const isCheckoutCompletedTrue = (url) => {
    const urlObj = new URL(url);
    const searchParams = urlObj.searchParams;
    let found = false;
    for (const value of searchParams.values()) {
        if (value === 'true') {
            return true;
        }
        if (value === 'false') {
            found = true;
        }
    }
    if (found) {
        return false;
    }
    return null;
};

export const convertLongNumbers = (num) => {
    if (num > 999999999) {
        return (num / 1000000000).toFixed(0) + 'B';
    } else if (num > 999999) {
        return (num / 1000000).toFixed(0) + 'M';
    } else if (num > 999) {
        return (num / 1000).toFixed(0) + 'K';
    } else return num;
};
