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

import typeIcon from '../../../../assets/images/typeIcon.svg';
import createdByIcon from '../../../../assets/images/createdByIcon.svg';
import { Add } from '@material-ui/icons';
import RadioButton from '../../../../components/radioButton';
import SelectVersion from '../../../../components/selectVersion';
import TagsList from '../../../../components/tagsList';

function SchemaDetails({ schema }) {
    const [passwordType, setPasswordType] = useState(0);
    const [versionSelected, setVersionSelected] = useState(`Version ${schema?.versions[0]?.label}`);

    const passwordOptions = [
        {
            id: 1,
            value: 0,
            label: 'Code'
        },
        {
            id: 2,
            value: 1,
            label: 'Table'
        }
    ];

    const passwordTypeChange = (e) => {
        setPasswordType(e.target.value);
    };

    const handleSelectVersion = (e) => {
        let index = schema.versions?.findIndex((version) => version.id === Number(e));
        setVersionSelected(`Version ${schema.versions[index].label}`);
    };

    return (
        <schema-details is="3xd">
            <div className="type-created">
                <div className="wrapper">
                    <img src={typeIcon} />
                    <p>Type:</p>
                    <span>{schema.type}</span>
                </div>
                <div className="wrapper">
                    <img src={createdByIcon} />
                    <p>Created by:</p>
                    <span>{schema.created_by}</span>
                </div>
            </div>
            <div className="tags">
                <TagsList tags={schema.tags} addNew={true} />
            </div>
            <div className="schema-fields">
                <div className="left">
                    <p>Schema</p>
                    <RadioButton options={passwordOptions} radioValue={passwordType} onChange={(e) => passwordTypeChange(e)} />
                </div>
                <SelectVersion value={versionSelected} options={schema.versions} onChange={(e) => handleSelectVersion(e)} />
            </div>
            <div className="schema-content"></div>
        </schema-details>
    );
}

export default SchemaDetails;
