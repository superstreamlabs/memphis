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

import { parsingDate } from '../../../../services/valueConvertor';
import sourceIcon from '../../../../assets/images/sourceIcon.svg';
import LogBadge from '../../../../components/logBadge';

const LogPayload = ({ value, onSelected, selectedRow }) => {
    return (
        <div className={selectedRow === value?.message_seq ? 'log-payload log-selected' : 'log-payload'} onClick={() => onSelected(value?.message_seq)}>
            {selectedRow === value?.message_seq && <div className="selected"></div>}
            <p className="title">{value?.data}</p>
            <p className="created-date">{parsingDate(value?.creation_date)}</p>
            <div className="log-info">
                <div className="source">
                    <img src={sourceIcon} alt="sourceIcon" />
                    <p>{value?.source}</p>
                </div>
                <LogBadge type={value?.type} />
            </div>
        </div>
    );
};

export default LogPayload;
