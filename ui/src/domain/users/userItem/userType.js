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

import React, { useEffect, useState } from 'react';

function UserType(props) {
    const [fontColor, setFontColor] = useState('#FD79A8');
    const [backgroundColor, setBackgroundColor] = useState('rgba(253, 121, 168, 0.2)');
    useEffect(() => {
        createTypeWrapperStyle(props.userType);
    }, []);

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
            {props.userType}
        </div>
    );
}

export default UserType;
