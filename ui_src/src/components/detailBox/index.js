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
import React from 'react';

const DetailBox = ({ img, title, desc, data }) => {
    return (
        <div className="detail-box-wrapper">
            <div className="detail-img">
                <img width={24} src={img} alt="leader" />
            </div>
            <div className="detail-title-wrapper">
                <div className="detail-title">{title}</div>
                <div className="detail-description">{desc}</div>
            </div>
            <div className="separator" />
            <div className="detail-data">
                {data.map((row) => {
                    return (
                        <div key={row} className="detail-data-row">
                            {row}
                        </div>
                    );
                })}
            </div>
        </div>
    );
};
export default DetailBox;
