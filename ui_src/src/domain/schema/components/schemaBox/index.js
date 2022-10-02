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
import usedIcond from '../../../../assets/images/usedIcon.svg';
import createdDateIcon from '../../../../assets/images/createdDateIcon.svg';
import notUsedIcond from '../../../../assets/images/notUsedIcon.svg';
import CheckboxComponent from '../../../../components/checkBox';
import { parsingDate } from '../../../../services/valueConvertor';
import Tag from '../../../../components/tag';
import { Drawer, Button, Space } from 'antd';
import SchemaDetails from '../schemaDetails';
import TagsList from '../../../../components/tagsList';

function SchemaBox({ schema, handleCheckedClick, isCheck }) {
    const [open, setOpen] = useState(false);
    const [size, setSize] = useState();

    const openSchemaDetails = () => {
        setSize('large');
        setOpen(true);
    };

    const onClose = () => {
        setOpen(false);
    };
    return (
        <>
            <div onClick={openSchemaDetails} key={schema.name} className="schema-box-wrapper">
                <header is="x3d">
                    <div className="header-wrapper">
                        <CheckboxComponent checked={isCheck} id={schema.id} onChange={handleCheckedClick} name={schema.id} />
                        <div className="schema-name">
                            <p>{schema.name}</p>
                        </div>
                        <div className="is-used">
                            <img src={schema.used ? usedIcond : notUsedIcond} />
                            {schema.used && <p className="used">Used</p>}
                            {!schema.used && <p className="not-used"> Not Used</p>}
                        </div>
                        <div className="menu">
                            <p>***</p>
                        </div>
                    </div>
                </header>
                <type is="x3d">
                    <div>
                        <p>Type : </p>
                        <span>{schema.type}</span>
                    </div>
                    <div>
                        <p>Created by : </p>
                        <span>{schema.created_by}</span>
                    </div>
                </type>
                <tags is="x3d">
                    <TagsList tags={schema.tags} />
                </tags>
                <date is="x3d">
                    <img src={createdDateIcon} />
                    <p>{parsingDate(schema.creation_date)}</p>
                </date>
            </div>
            <Drawer title={schema?.name} placement="right" size={'large'} onClose={onClose} open={open} maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}>
                <SchemaDetails schema={schema} />
            </Drawer>
        </>
    );
}

export default SchemaBox;
