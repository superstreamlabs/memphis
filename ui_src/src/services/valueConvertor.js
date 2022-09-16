// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

export const convertSecondsToDate = (seconds) => {
    const days = Math.floor(seconds / 86400);
    seconds -= days * 86400;
    const hours = Math.floor(seconds / 3600);
    seconds -= hours * 3600;
    const minutes = Math.floor(seconds / 60);
    seconds -= minutes * 60;

    let result = '';
    if (days > 0) {
        result = days === 1 ? 'one day' : `${days} days`;
        if (hours > 0) {
            result = hours === 1 ? `${result}, one hour` : `${result}, ${hours} hours`;
        }
        if (minutes > 0) {
            result = minutes === 1 ? `${result}, one minute` : `${result}, ${minutes} minutes`;
        }
        if (seconds > 0) {
            result = seconds === 1 ? `${result}, one second` : `${result}, ${seconds} seconds`;
        }
    } else if (hours > 0) {
        result = hours === 1 ? 'one hour' : `${hours} hours`;
        if (minutes > 0) {
            result = minutes === 1 ? `${result}, one minute` : `${result}, ${minutes} minutes`;
        }
        if (seconds > 0) {
            result = seconds === 1 ? `${result}, one second` : `${result}, ${seconds} seconds`;
        }
    } else if (minutes > 0) {
        result = minutes === 1 ? 'one minute' : `${minutes} minutes`;
        if (seconds > 0) {
            result = seconds === 1 ? `${result}, one second` : `${result}, ${seconds} seconds`;
        }
    } else if (seconds > 0) {
        result = seconds === 1 ? 'one second' : `${seconds} seconds`;
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
    return result;
};

export const parsingDate = (date) => {
    var options = { year: 'numeric', month: 'long', day: 'numeric', hour: 'numeric', minute: 'numeric' };
    return new Date(date).toLocaleDateString([], options);
};

function isFloat(n) {
    return Number(n) === n && n % 1 !== 0;
}

export const convertBytes = (bytes) => {
    const KB = 1024;
    const MB = 1024 * 1024;
    if (bytes < KB && bytes > 0) {
        return `${bytes} Bytes`;
    } else if (bytes >= KB && bytes < MB) {
        const parsing = isFloat(bytes / KB) ? Math.round((bytes / KB + Number.EPSILON) * 100) / 100 : bytes / KB;
        return `${parsing} KB`;
    } else if (bytes >= MB) {
        const parsing = isFloat(bytes / MB) ? Math.round((bytes / MB + Number.EPSILON) * 100) / 100 : bytes / MB;
        return `${parsing} MB`;
    } else {
        return '0 Bytes';
    }
};

export const numberWithCommas = (x) => {
    if (x) {
        return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    } else return 0;
};

export const capitalizeFirst = (str) => {
    return str?.charAt(0).toUpperCase() + str.slice(1);
};

export const cutInfoLog = (str) => {
    return str?.split('] ')[2];
};
