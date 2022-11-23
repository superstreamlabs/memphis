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

import React, { useEffect } from 'react';
import StatusIndication from '../../../../components/indication';

const Producer = ({ data }) => {
    const prod = data ? (
        <div className="poison-producer">
            <header is="x3d">
                <p>Producer</p>
                <StatusIndication is_active={data?.is_active} is_deleted={data?.is_deleted} />
            </header>
            <div className="content-wrapper">
                {data?.details?.length > 0 &&
                    data?.details?.map((row, index) => {
                        return (
                            <content is="x3d" key={index}>
                                <p>{row.name}</p>
                                <span>{row.value}</span>
                            </content>
                        );
                    })}
            </div>
        </div>
    ) : null;

    return <>{prod}</>;
};
export default Producer;
