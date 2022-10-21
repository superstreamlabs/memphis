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

import './style.scss';
import React, { useState } from 'react';
const ColorPicker = ({ onChange }) => {
    const colors = [
        '101, 87, 255',
        '77, 34, 178',
        '177, 140, 254',
        '216, 201, 254',
        '0, 165, 255',
        '238, 113, 158',
        '255, 140, 130',
        '252, 52, 0',
        '97, 223, 155',
        '32, 201, 172',
        '97, 223, 215',
        '255, 160, 67',
        '253, 236, 194',
        '182, 180, 186',
        '100, 100, 103'
    ];
    const [chosenColor, setChosenColor] = useState(colors[0]);

    const handleColorPick = (color) => {
        setChosenColor(color);
        onChange(color);
    };

    return colors?.map((color) => (
        <div className="color-picker">
            <li key={color} onClick={() => handleColorPick(color)}>
                <div className="color-circle" key={color} style={{ backgroundColor: `rgb(${color})` }}>
                    {color === chosenColor ? <div className="inner-circle"></div> : <></>}
                </div>
            </li>
        </div>
    ));
};

export default ColorPicker;
