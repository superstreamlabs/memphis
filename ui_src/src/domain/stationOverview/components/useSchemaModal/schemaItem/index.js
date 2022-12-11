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

import { BrokenImageRounded, CancelRounded, CloseRounded, FiberManualRecord, LinkOffRounded } from '@material-ui/icons';
import React, { useState } from 'react';

import schemaItemIcon from '../../../../../assets/images/schemaItemIcon.svg';
import deleteIcon from '../../../../../assets/images/deleteIcon.svg';

import { parsingDate } from '../../../../../services/valueConvertor';

const SchemaItem = ({ schema, schemaSelected, handleSelectedItem, selected, handleStopUseSchema }) => {
    return (
        <div
            key={schema?.id}
            className={selected === schema?.name ? 'schema-item-container sch-item-selected' : 'schema-item-container'}
            onClick={() => handleSelectedItem(schema?.name)}
        >
            <div className="content">
                <div className="name-wrapper">
                    <img src={schemaItemIcon} />
                    <p className="name">{schema?.name}</p>
                </div>
                <div className="details">
                    <p className="type">{schema?.type}</p>
                    <FiberManualRecord />
                    <p className="date">{parsingDate(schema?.creation_date)}</p>
                </div>
            </div>
            {schema?.name === schemaSelected && (
                <div className="delete-icon" onClick={handleStopUseSchema}>
                    <CloseRounded />
                </div>
            )}
        </div>
    );
};

export default SchemaItem;
