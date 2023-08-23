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

export const generator = () => {
    const uppercase = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const lowercase = 'abcdefghijklmnopqrstuvwxyz';
    const numeric = '0123456789';
    const specialChars = '!?-@#$%';
    const minLength = 10;
    const maxLength = 20;

    let password = '';

    const getRandomChar = (characters) => {
        const randomIndex = Math.floor(Math.random() * characters.length);
        return characters[randomIndex];
    };

    const hasUppercase = /(?=.*[A-Z])/;
    const hasLowercase = /(?=.*[a-z])/;
    const hasNumeric = /(?=.*\d)/;
    const hasSpecialChars = /(?=.*[!?\-@#$%])/;

    while (
        (!hasUppercase.test(password) || !hasLowercase.test(password) || !hasNumeric.test(password) || !hasSpecialChars.test(password) || password.length < minLength) &&
        password.length < maxLength
    ) {
        const charSetIndex = Math.floor(Math.random() * 4);
        let randomChar;

        switch (charSetIndex) {
            case 0:
                randomChar = getRandomChar(uppercase);
                break;
            case 1:
                randomChar = getRandomChar(lowercase);
                break;
            case 2:
                randomChar = getRandomChar(numeric);
                break;
            case 3:
                randomChar = getRandomChar(specialChars);
                break;
            default:
                break;
        }

        password += randomChar;
    }

    password = password
        .split('')
        .sort(() => Math.random() - 0.5)
        .join('');

    return password;
};
