// Credit for The NATS.IO Authors
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

import './style.scss';

import React, { useContext, useEffect, useState } from 'react';

import { Context } from '../../../../../hooks/store';
import { Slider } from 'antd';

function SliderRow({ title, desc, value, onChanges, img, min, max, unit }) {
    const [inputValue, setInputValue] = useState(value);
    const [state, dispatch] = useContext(Context);

    const onChange = (newValue) => {
        setInputValue(newValue);
        onChanges(newValue);
    };
    useEffect(() => {
        setInputValue(value);
    }, [value]);

    return (
        <div className="configuration-list-container">
            <div className="left-side">
                <img src={img} alt="ConfImg2" />
                <div>
                    <p className="conf-name">{title}</p>
                    <label className="conf-description">{desc}</label>
                </div>
            </div>
            <div className="current-value">
                <p>
                    {inputValue} {unit}
                </p>
            </div>
            <div className="right-side">
                <div className="min-max-box">
                    <span>
                        {min} {unit}
                    </span>
                </div>
                <Slider
                    style={{ width: '20vw' }}
                    min={min}
                    max={max}
                    onChange={onChange}
                    value={inputValue}
                    trackStyle={{ background: 'var(--purple)', height: '4px' }}
                    handleStyle={{
                        border: '8px solid #FFFFFF',
                        background: 'var(--purple)',
                        boxShadow: '0px 8px 16px rgba(0, 82, 204, 0.16)',
                        width: '24px',
                        height: '24px',
                        marginTop: '-10px'
                    }}
                />
                <div className="min-max-box">
                    <span>
                        {max} {unit}
                    </span>
                </div>
            </div>
        </div>
    );
}

export default SliderRow;
