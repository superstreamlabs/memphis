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

import { Drawer, Button, Space } from 'antd';
import React, { useState } from 'react';

import createdDateIcon from '../../../../assets/images/createdDateIcon.svg';
import notUsedIcond from '../../../../assets/images/notUsedIcon.svg';
import { parsingDate } from '../../../../services/valueConvertor';
import CheckboxComponent from '../../../../components/checkBox';
import usedIcond from '../../../../assets/images/usedIcon.svg';
import TagsList from '../../../../components/tagsList';
import Tag from '../../../../components/tag';
import SchemaDetails from '../schemaDetails';

const tags = [
    { name: 'Github', background: 'rgba(0, 165, 255,0.1)', color: 'rgb(0, 165, 255)' },
    { name: 'R&D', background: 'rgba(32, 201, 172,0.1)', color: '#20C9AC' },
    { name: 'Avi_team', background: 'rgba(85, 66, 246, 0.1)', color: 'rgba(85, 66, 246, 1)' },
    { name: 'Logs', background: 'rgba(252, 52, 0, 0.1)', color: 'rgba(252, 52, 0, 1)' }
];

function SchemaBox({ schema, handleCheckedClick, isCheck }) {
    const [open, setOpen] = useState(false);
    const [size, setSize] = useState();

    const openSchemaDetails = () => {
        setSize('large');
        setOpen(true);
    };

    const closeDrawer = () => {
        setOpen(false);
    };
    return (
        <>
            <div key={schema.name} className="schema-box-wrapper">
                <header is="x3d">
                    <div className="header-wrapper">
                        <CheckboxComponent checked={isCheck} id={schema.name} onChange={handleCheckedClick} name={schema.name} />
                        <div onClick={openSchemaDetails} className="schema-name">
                            <p>{schema.name}</p>
                        </div>
                        <div onClick={openSchemaDetails} className="is-used">
                            <img src={schema.used ? usedIcond : notUsedIcond} />
                            {schema.used && <p className="used">Used</p>}
                            {!schema.used && <p className="not-used"> Not Used</p>}
                        </div>
                        {/* <div className="menu">
                            <p>***</p>
                        </div> */}
                    </div>
                </header>
                <type onClick={openSchemaDetails} is="x3d">
                    <div>
                        <p>Type : </p>
                        <span>{schema.type}</span>
                    </div>
                    <div>
                        <p>Created by : </p>
                        <span>{schema.created_by_user}</span>
                    </div>
                </type>
                <tags onClick={openSchemaDetails} is="x3d">
                    <TagsList tags={tags} />
                </tags>
                <date onClick={openSchemaDetails} is="x3d">
                    <img src={createdDateIcon} />
                    <p>{parsingDate(schema.creation_date)}</p>
                </date>
            </div>
            <Drawer title={schema?.name} placement="right" size={'large'} onClose={closeDrawer} open={open} maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}>
                <SchemaDetails schemaName={schema?.name} closeDrawer={() => closeDrawer()} />
            </Drawer>
        </>
    );
}

export default SchemaBox;
