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

import { CheckBox } from '@material-ui/icons';
import React from 'react';
import usedIcond from '../../../../assets/images/usedIcon.svg';
import notUsedIcond from '../../../../assets/images/notUsedIcon.svg';

function SchemaBox() {
    return (
        <div className="schema-box-wrapper">
            <header is="x3d">
                <CheckBox />
                <div className="schema-id">
                    <p>1D1F1R2T3W</p>
                </div>
                <div className="is-used">
                    <img src={usedIcond} />
                </div>
                <div className="menu">
                    <p>***</p>
                </div>
            </header>
            <title is="x3d"></title>
        </div>
    );
}

export default SchemaBox;
