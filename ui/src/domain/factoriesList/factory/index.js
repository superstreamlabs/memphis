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

import React, { useState, useContext, useEffect } from 'react';
import MoreVertIcon from '@material-ui/icons/MoreVert';
import DeleteOutline from '@material-ui/icons/DeleteOutline';
import { Link } from 'react-router-dom';
import MenuItem from '@material-ui/core/MenuItem';
import Popover from '@material-ui/core/Popover';
import EditOutlined from '@material-ui/icons/EditOutlined';
import { Divider } from '@material-ui/core';

import OverflowTip from '../../../components/tooltip/overflowtip';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import Modal from '../../../components/modal';
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';
import { parsingDate } from '../../../services/valueConvertor';

const Factory = (props) => {
    const [state, dispatch] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [anchorEl, setAnchorEl] = useState(null);
    const open = Boolean(anchorEl);
    const [botUrl, SetBotUrl] = useState('');
    const botId = 1;

    useEffect(() => {
        setBotImage(props.content?.user_avatar_id || botId);
    }, []);

    const setBotImage = (botId) => {
        SetBotUrl(require(`../../../assets/images/bots/${botId}.svg`));
    };

    const handleClickMenu = (event) => {
        setAnchorEl(event.currentTarget);
    };

    const handleCloseMenu = () => {
        setAnchorEl(null);
    };

    const removeFactory = async () => {
        try {
            await httpRequest('DELETE', ApiEndpoints.REMOVE_FACTORY, {
                factory_name: props.content.name
            });
            modalFlip(false);
            props.removeFactory();
        } catch (err) {}
    };

    return (
        <div className="factory">
            <div className="factory-card-container" key={props.content.id}>
                <Link style={{ cursor: 'pointer' }} to={`${pathDomains.factoriesList}/${props.content.name}`}>
                    <div className="factory-card-title">
                        <h2>
                            <OverflowTip text={props.content.name} width={'220px'} color="white" cursor="pointer">
                                {props.content.name}
                            </OverflowTip>
                        </h2>
                        <div id="e2e-tests-factoty-open">
                            <MoreVertIcon
                                aria-controls="long-button"
                                aria-haspopup="true"
                                onClick={(e) => {
                                    e.preventDefault();
                                    handleClickMenu(e);
                                }}
                                className="threedots-menu"
                            />
                        </div>
                    </div>
                    <div className="factory-card-description">
                        <p>{props.content.description || 'Empty description'}</p>
                    </div>
                </Link>
                <div className="factory-owner">
                    <div className="user-avatar">
                        <img className="sandboxUserImg" src={botUrl} width={25} height={25} alt="bot"></img>
                    </div>
                    <div className="user-details">
                        <p>{props.content.created_by_user}</p>
                        <span>{parsingDate(props.content.creation_date)}</span>
                    </div>
                </div>
                <Popover id="long-menu" classes={{ paper: 'Menu' }} anchorEl={anchorEl} onClose={handleCloseMenu} open={open}>
                    <Link to={`${pathDomains.factoriesList}/${props.content.name}`}>
                        <MenuItem
                            onClick={() => {
                                handleCloseMenu();
                            }}
                        >
                            <EditOutlined className="menu-item-icon" />
                            <label className="menu-item-label">Edit</label>
                        </MenuItem>
                    </Link>
                    <Divider />
                    <MenuItem
                        onClick={() => {
                            modalFlip(true);
                            handleCloseMenu();
                        }}
                    >
                        <DeleteOutline className="menu-item-icon" />
                        <label id="e2e-tests-factoty-remove" className="menu-item-label">
                            Remove
                        </label>
                    </MenuItem>
                </Popover>
            </div>
            <Modal
                header="Remove Factory"
                height="260px"
                minWidth="460px"
                rBtnText="Cancel"
                lBtnText="Remove"
                closeAction={() => modalFlip(false)}
                lBtnClick={() => {
                    removeFactory();
                }}
                clickOutside={() => modalFlip(false)}
                rBtnClick={() => {
                    modalFlip(false);
                }}
                open={modalIsOpen}
            >
                Are you sure you want to remove "<b>{props.content.name}</b>" factory? This will remove all stations in this factory.
            </Modal>
        </div>
    );
};

export default Factory;
