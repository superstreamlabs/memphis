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

import { CloseRounded } from '@material-ui/icons';
import { Drawer, Button, Space } from 'antd';
import React, { useState } from 'react';

import createdDateIcon from '../../../../assets/images/createdDateIcon.svg';
import notUsedIcond from '../../../../assets/images/notUsedIcon.svg';
import { parsingDate } from '../../../../services/valueConvertor';
import CheckboxComponent from '../../../../components/checkBox';
import usedIcond from '../../../../assets/images/usedIcon.svg';
import TagsList from '../../../../components/tagList';
import SchemaDetails from '../schemaDetails';

function SchemaBox({ schema, handleCheckedClick, isCheck }) {
    const [open, setOpen] = useState(false);

    const handleDrawer = (flag) => {
        setOpen(flag);
    };

    return (
        <>
            <div>
                <CheckboxComponent checked={isCheck} id={schema.name} onChange={handleCheckedClick} name={schema.name} />
                <div key={schema.name} onClick={() => handleDrawer(true)} className="schema-box-wrapper">
                    <header is="x3d">
                        <div className="header-wrapper">
                            <div className="schema-name">
                                <p>{schema.name}</p>
                            </div>
                            <div className="is-used">
                                <img src={schema.used ? usedIcond : notUsedIcond} alt="usedIcond" />
                                {schema.used && <p className="used">Used</p>}
                                {!schema.used && <p className="not-used"> Not used</p>}
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
                            <span>{schema.created_by_user}</span>
                        </div>
                    </type>
                    <tags is="x3d">
                        <TagsList tagsToShow={3} tags={schema?.tags} />
                    </tags>
                    <date is="x3d">
                        <img src={createdDateIcon} alt="createdDateIcon" />
                        <p>{parsingDate(schema.creation_date)}</p>
                    </date>
                </div>
            </div>
            <Drawer
                title={schema?.name}
                placement="right"
                size={'large'}
                onClose={() => handleDrawer(false)}
                open={open}
                maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}
                closeIcon={<CloseRounded style={{ color: '#D1D1D1' }} />}
            >
                <SchemaDetails schemaName={schema?.name} closeDrawer={() => handleDrawer(false)} />
            </Drawer>
        </>
    );
}

export default SchemaBox;
