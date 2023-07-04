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
    const length = 10;

    let password = '';
    let character = '';

    while (password.length < length) {
        const entity1 = Math.floor(Math.random() * uppercase.length);
        const entity2 = Math.floor(Math.random() * lowercase.length);
        const entity3 = Math.floor(Math.random() * numeric.length);
        const entity4 = Math.floor(Math.random() * specialChars.length);

        character += uppercase.charAt(entity1);
        character += lowercase.charAt(entity2);
        character += numeric.charAt(entity3);
        character += specialChars.charAt(entity4);

        password = character;
    }

    password = password
        .split('')
        .sort(() => {
            return 0.5 - Math.random();
        })
        .join('');

    return password.substr(0, length);
};
