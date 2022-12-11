// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import React, { useEffect, useState } from 'react';

function UserType({ userType }) {
    const [fontColor, setFontColor] = useState('#FD79A8');
    const [backgroundColor, setBackgroundColor] = useState('rgba(253, 121, 168, 0.2)');

    useEffect(() => {
        createTypeWrapperStyle(userType);
    }, [userType]);

    const createTypeWrapperStyle = (userType) => {
        switch (userType) {
            case 'management':
                setFontColor('#36DEDE');
                setBackgroundColor('rgba(54, 222, 222, 0.2)');
                break;
            case 'application':
                setFontColor('#419FFE');
                setBackgroundColor('rgba(65, 159, 254, 0.2)');
                break;
            default:
                setFontColor('#FD79A8');
                setBackgroundColor('rgba(253, 121, 168, 0.2))');
                break;
        }
    };

    return (
        <div className="user-typep-wrapper" style={{ background: backgroundColor, color: fontColor }}>
            {userType}
        </div>
    );
}

export default UserType;
